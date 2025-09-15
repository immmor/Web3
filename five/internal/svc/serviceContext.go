package svc

import (
	"five/internal/config"
	"five/internal/middleware"
	"five/internal/types"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/rest"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config    config.Config
	Auth      rest.Middleware
	Log       rest.Middleware
	MySQL     *gorm.DB
	Redis     *redis.Client
	KafkaProd *kafka.Writer
	KafkaCons *kafka.Reader
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化MySQL
	db, err := gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移数据库表 - 先删除表再重新创建
	db.Exec("DROP TABLE IF EXISTS trades")
	db.Exec("DROP TABLE IF EXISTS orders")
	err = db.AutoMigrate(&types.Order{}, &types.Trade{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// 初始化Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	// 初始化Kafka生产者
	producer := &kafka.Writer{
		Addr:     kafka.TCP(c.Kafka.Brokers...),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}

	// 初始化Kafka消费者
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Kafka.Brokers,
		Topic:    "orders",
		GroupID:  "order-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &ServiceContext{
		Config:    c,
		Auth:      middleware.NewAuthMiddleware().Handle,
		Log:       middleware.NewLogMiddleware().Handle,
		MySQL:     db,
		Redis:     rdb,
		KafkaProd: producer,
		KafkaCons: consumer,
	}
}

package types

import (
	"time"
)

// 订单类型
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"  // 限价单
	OrderTypeMarket OrderType = "market" // 市价单
)

// 订单方向
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"  // 买入
	OrderSideSell OrderSide = "sell" // 卖出
)

// 订单状态
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"    // 挂单中
	OrderStatusPartFilled OrderStatus = "part_filled" // 部分成交
	OrderStatusFilled     OrderStatus = "filled"      // 完全成交
	OrderStatusCancelled  OrderStatus = "cancelled"   // 已取消
	OrderStatusRejected   OrderStatus = "rejected"    // 已拒绝
)

type Order struct {
	ID              uint        `gorm:"primaryKey;autoIncrement" json:"-"`
	CreatedAt       time.Time   `gorm:"autoCreateTime" json:"-"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime" json:"-"`
	OrderID         string      `gorm:"size:100;uniqueIndex" json:"order_id"`
	UserID          int64       `json:"user_id"`
	Symbol          string      `gorm:"size:20;not null" json:"symbol"`           // 交易对，如 BTC/USDT
	OrderType       OrderType   `gorm:"size:20;not null" json:"order_type"`        // 订单类型：limit, market
	OrderSide       OrderSide   `gorm:"size:10;not null" json:"order_side"`        // 订单方向：buy, sell
	Price           float64     `gorm:"not null;default:0" json:"price"`           // 限价单价格
	Amount          float64     `gorm:"not null;default:0" json:"amount"`          // 订单数量
	FilledAmount    float64     `gorm:"not null;default:0" json:"filled_amount"`    // 已成交数量
	Fee             float64     `gorm:"not null;default:0" json:"fee"`              // 手续费
	FeeAsset        string      `gorm:"size:10;default:'USDT'" json:"fee_asset"`   // 手续费币种
	Status          OrderStatus `gorm:"size:20;not null;default:'pending'" json:"status"` // 订单状态
	CancelReason    string      `gorm:"size:200;default:''" json:"cancel_reason"`  // 取消原因
}

type OrderMessage struct {
	Action  string `json:"action"` // create, update, cancel, fill
	OrderID string `json:"order_id"`
	Data    Order  `json:"data"`
}

// 成交记录
type Trade struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	TradeID   string    `gorm:"size:100;uniqueIndex" json:"trade_id"`
	OrderID   string    `gorm:"size:100;index" json:"order_id"`
	Symbol    string    `gorm:"size:20;not null" json:"symbol"`
	Price     float64   `gorm:"not null;default:0" json:"price"`
	Amount    float64   `gorm:"not null;default:0" json:"amount"`
	Fee       float64   `gorm:"not null;default:0" json:"fee"`
	FeeAsset  string    `gorm:"size:10;default:'USDT'" json:"fee_asset"`
}
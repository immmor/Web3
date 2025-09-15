package order

import (
	"context"
	"encoding/json"
	"five/internal/svc"
	"five/internal/types"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateOrder 创建订单（挂单）
func (l *OrderLogic) CreateOrder(order *types.Order) error {
	// 确保所有字段都有值
	if order.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if order.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if order.Symbol == "" {
		order.Symbol = "BTC/USDT"
	}
	if order.OrderType == "" {
		order.OrderType = types.OrderTypeLimit
	}
	if order.OrderSide == "" {
		order.OrderSide = types.OrderSideBuy
	}
	if order.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if order.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	
	// 设置默认值
	order.FilledAmount = 0
	order.Fee = 0
	order.FeeAsset = "USDT"
	order.Status = types.OrderStatusPending
	order.CancelReason = ""

	// 1. 写入MySQL - 直接使用结构体，让GORM处理字段映射
	if err := l.svcCtx.MySQL.Create(order).Error; err != nil {
		return err
	}

	// 2. 写入Redis缓存
	if err := l.updateOrderCache(order); err != nil {
		return err
	}

	// 3. 发送Kafka消息
	return l.sendOrderMessage("create", order)
}

// CancelOrder 取消订单（下架）
func (l *OrderLogic) CancelOrder(orderID, reason string) error {
	// 1. 查询订单
	order, err := l.GetOrder(orderID)
	if err != nil {
		return err
	}

	// 检查订单状态是否可以取消
	if order.Status != types.OrderStatusPending && order.Status != types.OrderStatusPartFilled {
		return fmt.Errorf("订单状态 %s 不可取消", order.Status)
	}

	// 2. 更新订单状态
	order.Status = types.OrderStatusCancelled
	order.CancelReason = reason
	order.UpdatedAt = time.Now()

	// 3. 更新数据库 - 使用结构体更新，让GORM处理字段映射
	if err := l.svcCtx.MySQL.Model(order).
		Where("order_id = ?", orderID).
		Updates(order).Error; err != nil {
		return err
	}

	// 4. 更新缓存
	if err := l.updateOrderCache(order); err != nil {
		return err
	}

	// 5. 发送取消消息
	return l.sendOrderMessage("cancel", order)
}

// FillOrder 订单成交（部分或完全）
func (l *OrderLogic) FillOrder(orderID string, fillPrice, fillAmount, feeRate float64) error {
	// 1. 查询订单
	order, err := l.GetOrder(orderID)
	if err != nil {
		return err
	}

	// 计算手续费
	fee := fillAmount * fillPrice * feeRate
	order.FilledAmount += fillAmount
	order.Fee += fee

	// 更新订单状态
	if order.FilledAmount >= order.Amount {
		order.Status = types.OrderStatusFilled
	} else if order.FilledAmount > 0 {
		order.Status = types.OrderStatusPartFilled
	}
	order.UpdatedAt = time.Now()

	// 2. 更新数据库 - 使用结构体更新，让GORM处理字段映射
	if err := l.svcCtx.MySQL.Model(order).
		Where("order_id = ?", orderID).
		Updates(order).Error; err != nil {
		return err
	}

	// 3. 创建成交记录
	trade := &types.Trade{
		TradeID:  fmt.Sprintf("trade_%d", time.Now().UnixNano()),
		OrderID:  orderID,
		Symbol:   order.Symbol,
		Price:    fillPrice,
		Amount:   fillAmount,
		Fee:      fee,
		FeeAsset: "USDT", // 假设手续费币种
	}

	if err := l.svcCtx.MySQL.Create(trade).Error; err != nil {
		return err
	}

	// 4. 更新缓存
	if err := l.updateOrderCache(order); err != nil {
		return err
	}

	// 5. 发送成交消息
	return l.sendOrderMessage("fill", order)
}

// GetOrder 获取订单（先查Redis缓存，没有再查MySQL）
func (l *OrderLogic) GetOrder(orderID string) (*types.Order, error) {
	// 1. 先查Redis缓存
	orderKey := fmt.Sprintf("order:%s", orderID)
	cachedOrder, err := l.svcCtx.Redis.Get(l.ctx, orderKey).Bytes()
	if err == nil {
		var order types.Order
		if json.Unmarshal(cachedOrder, &order) == nil {
			return &order, nil
		}
	}

	// 2. 缓存未命中，查MySQL
	var order types.Order
	if err := l.svcCtx.MySQL.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	// 3. 回写缓存
	if err := l.updateOrderCache(&order); err != nil {
		return &order, err // 返回订单，即使缓存更新失败
	}

	return &order, nil
}

// updateOrderCache 更新订单缓存
func (l *OrderLogic) updateOrderCache(order *types.Order) error {
	orderKey := fmt.Sprintf("order:%s", order.OrderID)
	orderJSON, _ := json.Marshal(order)
	return l.svcCtx.Redis.Set(l.ctx, orderKey, orderJSON, 24*time.Hour).Err()
}

// sendOrderMessage 发送Kafka消息
func (l *OrderLogic) sendOrderMessage(action string, order *types.Order) error {
	message := types.OrderMessage{
		Action:  action,
		OrderID: order.OrderID,
		Data:    *order,
	}
	msgJSON, _ := json.Marshal(message)

	return l.svcCtx.KafkaProd.WriteMessages(l.ctx, kafka.Message{
		Key:   []byte(order.OrderID),
		Value: msgJSON,
	})
}

// Kafka消费者处理消息
func (l *OrderLogic) StartKafkaConsumer() {
	go func() {
		for {
			msg, err := l.svcCtx.KafkaCons.ReadMessage(l.ctx)
			if err != nil {
				fmt.Printf("Kafka消费错误: %v\n", err)
				continue
			}

			var orderMsg types.OrderMessage
			if err := json.Unmarshal(msg.Value, &orderMsg); err != nil {
				fmt.Printf("消息解析错误: %v\n", err)
				continue
			}

			switch orderMsg.Action {
			case "create":
				fmt.Printf("订单创建: %s, 状态: %s\n", orderMsg.OrderID, orderMsg.Data.Status)
			case "cancel":
				fmt.Printf("订单取消: %s, 原因: %s\n", orderMsg.OrderID, orderMsg.Data.CancelReason)
			case "fill":
				fmt.Printf("订单成交: %s, 成交数量: %.4f, 手续费: %.4f\n",
					orderMsg.OrderID, orderMsg.Data.FilledAmount, orderMsg.Data.Fee)
			}
		}
	}()
}

// 获取用户的所有订单
func (l *OrderLogic) GetUserOrders(userID int64, status types.OrderStatus) ([]types.Order, error) {
	var orders []types.Order
	query := l.svcCtx.MySQL.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// 获取订单的成交记录
func (l *OrderLogic) GetOrderTrades(orderID string) ([]types.Trade, error) {
	var trades []types.Trade
	if err := l.svcCtx.MySQL.Where("order_id = ?", orderID).Order("created_at ASC").Find(&trades).Error; err != nil {
		return nil, err
	}
	return trades, nil
}

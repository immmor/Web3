#!/bin/bash

# 完整的订单生命周期测试脚本
echo "=== 订单生命周期完整测试 ==="

# 1. 创建订单
echo "1. 创建订单..."
curl -X POST "http://localhost:8888/order/create" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "test_order_001",
    "user_id": 123,
    "symbol": "BTC/USDT",
    "order_type": "limit",
    "order_side": "buy",
    "price": 50000.0,
    "amount": 1.0
  }'

echo -e "\n\n2. 查询订单..."
curl "http://localhost:8888/order/get?order_id=test_order_001"

echo -e "\n\n3. 查询用户订单..."
curl "http://localhost:8888/order/user-orders?user_id=123"

echo -e "\n\n4. 部分成交订单 (0.5)..."
curl -X POST "http://localhost:8888/order/fill?order_id=test_order_001&price=50000.0&amount=0.5&fee_rate=0.001"

echo -e "\n\n5. 查询订单状态（部分成交）..."
curl "http://localhost:8888/order/get?order_id=test_order_001"

echo -e "\n\n6. 查询成交记录..."
curl "http://localhost:8888/order/trades?order_id=test_order_001"

echo -e "\n\n7. 完全成交订单 (剩余0.5)..."
curl -X POST "http://localhost:8888/order/fill?order_id=test_order_001&price=50000.0&amount=0.5&fee_rate=0.001"

echo -e "\n\n8. 查询订单状态（完全成交）..."
curl "http://localhost:8888/order/get?order_id=test_order_001"

echo -e "\n\n9. 尝试取消已成交订单（应该失败）..."
curl -X POST "http://localhost:8888/order/cancel?order_id=test_order_001&reason=test_cancel"

echo -e "\n\n10. 创建新订单用于取消测试..."
curl -X POST "http://localhost:8888/order/create" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "test_order_002",
    "user_id": 123,
    "symbol": "ETH/USDT",
    "order_type": "limit",
    "order_side": "sell",
    "price": 3000.0,
    "amount": 10.0
  }'

echo -e "\n\n11. 取消订单..."
curl -X POST "http://localhost:8888/order/cancel?order_id=test_order_002&reason=user_request"

echo -e "\n\n12. 查询取消后的订单状态..."
curl "http://localhost:8888/order/get?order_id=test_order_002"

echo -e "\n\n13. 查询用户所有订单..."
curl "http://localhost:8888/order/user-orders?user_id=123"

echo -e "\n\n=== 测试完成 ==="

# 数据库检查
echo -e "\n=== 数据库检查 ==="
docker exec -it mysql mysql -uapp_user -papp_password app_db -e "SELECT order_id, symbol, order_type, order_side, price, amount, filled_amount, status FROM orders;"

echo -e "\n=== 成交记录检查 ==="
docker exec -it mysql mysql -uapp_user -papp_password app_db -e "SELECT order_id, price, amount, fee, fee_asset FROM trades;"

echo -e "\n=== Redis缓存检查 ==="
docker exec -it redis redis-cli KEYS "order:*"
#!/bin/bash

echo "=========================================="
echo "测试活动状态自动更新功能"
echo "=========================================="

# 检查admin服务是否运行
if ! curl -s http://localhost:8081/api/health > /dev/null 2>&1; then
    echo "❌ Admin服务未运行，请先启动: go run ./cmd/admin"
    exit 1
fi

echo ""
echo "1. 测试活动列表API..."
echo "   请求: GET /api/seckill-activities"
response=$(curl -s http://localhost:8081/api/seckill-activities)
echo "$response" | jq -r '.data[] | "   活动ID: \(.ID), 名称: \(.Name), 状态: \(.Status), 结束时间: \(.EndTime)"' 2>/dev/null || echo "$response"

echo ""
echo "2. 测试商品列表API..."
echo "   请求: GET /api/products"
response2=$(curl -s http://localhost:8081/api/products)
echo "$response2" | jq -r '.data[] | select(.Status == 2) | "   商品ID: \(.ID), 名称: \(.Name), 状态: \(.Status), 秒杀库存: \(.SeckillStock)"' 2>/dev/null || echo "$response2"

echo ""
echo "3. 检查是否有状态为2（秒杀中）的商品..."
seckill_count=$(echo "$response2" | jq '[.data[] | select(.Status == 2)] | length' 2>/dev/null || echo "0")
if [ "$seckill_count" = "0" ]; then
    echo "   ✅ 没有处于秒杀状态的商品（符合预期）"
else
    echo "   ⚠️  仍有 $seckill_count 个商品处于秒杀状态"
fi

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo ""
echo "提示：如果活动已过期但状态仍显示'进行中'，请："
echo "1. 确认admin服务已重启"
echo "2. 刷新后台管理页面"
echo "3. 检查活动结束时间是否正确"

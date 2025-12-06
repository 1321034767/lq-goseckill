#!/bin/bash

echo "测试路由..."
echo ""

echo "1. 测试 /api/products/1/seckill-stock"
curl -s http://localhost:8080/api/products/1/seckill-stock | jq . || echo "失败"
echo ""

echo "2. 测试 /api/monitor/stats (admin)"
curl -s http://localhost:8081/api/monitor/stats | jq . || echo "失败"
echo ""

echo "3. 测试登录获取token"
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}' | jq -r '.data.token // empty')
echo "Token: $TOKEN"
echo ""

if [ -n "$TOKEN" ]; then
  echo "4. 测试 /api/seckill/1/result (需要token)"
  curl -s http://localhost:8080/api/seckill/1/result \
    -H "Authorization: $TOKEN" | jq . || echo "失败"
  echo ""
fi

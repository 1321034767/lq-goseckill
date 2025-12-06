#!/bin/bash

echo "=========================================="
echo "    重启服务并运行测试"
echo "=========================================="
echo ""

# 停止现有服务
echo "1. 停止现有服务..."
pkill -f "cmd/web" 2>/dev/null
pkill -f "cmd/admin" 2>/dev/null
pkill -f "cmd/seckill-worker" 2>/dev/null
sleep 2
echo "   完成"

# 重新启动服务
echo ""
echo "2. 启动服务..."
cd /root/internship-preparation/goseckill-12-01

nohup go run ./cmd/web > web.log 2>&1 &
WEB_PID=$!
echo "   Web服务启动 (PID: $WEB_PID)"

nohup go run ./cmd/admin > admin.log 2>&1 &
ADMIN_PID=$!
echo "   Admin服务启动 (PID: $ADMIN_PID)"

nohup go run ./cmd/seckill-worker > worker.log 2>&1 &
WORKER_PID=$!
echo "   Worker服务启动 (PID: $WORKER_PID)"

# 等待服务启动
echo ""
echo "3. 等待服务启动..."
sleep 5

# 检查服务状态
echo ""
echo "4. 检查服务状态..."
WEB_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health 2>/dev/null || echo "000")
ADMIN_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/api/health 2>/dev/null || echo "000")

if [ "$WEB_STATUS" = "200" ]; then
    echo "   ✅ Web服务运行正常 (8080)"
else
    echo "   ❌ Web服务未响应 (8080)"
fi

if [ "$ADMIN_STATUS" = "200" ]; then
    echo "   ✅ Admin服务运行正常 (8081)"
else
    echo "   ❌ Admin服务未响应 (8081)"
fi

# 运行测试
if [ "$WEB_STATUS" = "200" ]; then
    echo ""
    echo "5. 运行测试程序..."
    echo ""
    cd cmd/test-seckill-features
    go run main.go
else
    echo ""
    echo "⚠️  Web服务未启动，无法运行测试"
    echo "   请检查日志: tail -f web.log"
fi

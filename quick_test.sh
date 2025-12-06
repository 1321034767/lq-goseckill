#!/bin/bash

# 快速测试脚本

echo "=========================================="
echo "    秒杀功能快速测试"
echo "=========================================="
echo ""

# 检查服务
echo "【步骤1】检查服务状态..."
WEB_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health 2>/dev/null || echo "000")
ADMIN_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/api/health 2>/dev/null || echo "000")

if [ "$WEB_STATUS" = "200" ]; then
    echo "✅ Web服务运行正常 (8080)"
else
    echo "❌ Web服务未运行 (8080)"
    echo "   请先启动: go run ./cmd/web &"
fi

if [ "$ADMIN_STATUS" = "200" ]; then
    echo "✅ Admin服务运行正常 (8081)"
else
    echo "❌ Admin服务未运行 (8081)"
    echo "   请先启动: go run ./cmd/admin &"
fi

echo ""

# 运行测试程序
if [ "$WEB_STATUS" = "200" ]; then
    echo "【步骤2】运行功能测试..."
    echo ""
    cd cmd/test-seckill-features
    go run main.go
else
    echo "⚠️  请先启动Web和Admin服务后再运行测试"
    echo ""
    echo "启动命令："
    echo "  go run ./cmd/web &"
    echo "  go run ./cmd/admin &"
    echo "  go run ./cmd/seckill-worker &"
    echo ""
    echo "然后运行："
    echo "  bash quick_test.sh"
fi

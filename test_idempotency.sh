#!/bin/bash

# 幂等性测试脚本
# 用于快速测试秒杀系统的幂等性保证功能

set -e

echo "=========================================="
echo "    幂等性测试脚本"
echo "=========================================="
echo ""

# 检查服务是否运行
check_service() {
    local service=$1
    local port=$2
    
    if ! nc -z localhost $port 2>/dev/null; then
        echo "❌ 错误: $service 服务未运行 (端口 $port)"
        echo "   请先启动服务:"
        echo "   go run ./cmd/web &"
        echo "   go run ./cmd/seckill-worker &"
        exit 1
    fi
    echo "✅ $service 服务运行正常 (端口 $port)"
}

echo "步骤1: 检查服务状态..."
check_service "Web" 8080
check_service "MySQL" 3306
check_service "Redis" 6379
check_service "RabbitMQ" 5672
echo ""

# 检查Worker是否运行（通过检查进程）
if ! pgrep -f "seckill-worker" > /dev/null; then
    echo "⚠️  警告: seckill-worker 进程未运行"
    echo "   建议启动: go run ./cmd/seckill-worker &"
    echo ""
fi

echo "步骤2: 运行幂等性测试程序..."
echo ""

cd cmd/test-idempotency
go run main.go

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="

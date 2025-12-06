#!/bin/bash

# 秒杀功能完整测试脚本

set -e

echo "=========================================="
echo "    秒杀功能完整测试脚本"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查服务是否运行
check_service() {
    local service=$1
    local port=$2
    
    if ! nc -z localhost $port 2>/dev/null; then
        echo -e "${RED}❌ 错误: $service 服务未运行 (端口 $port)${NC}"
        echo "   请先启动服务:"
        echo "   go run ./cmd/web &"
        echo "   go run ./cmd/admin &"
        echo "   go run ./cmd/seckill-worker &"
        exit 1
    fi
    echo -e "${GREEN}✅ $service 服务运行正常 (端口 $port)${NC}"
}

echo "步骤1: 检查服务状态..."
check_service "Web" 8080
check_service "Admin" 8081
check_service "MySQL" 3306
check_service "Redis" 6379
check_service "RabbitMQ" 5672
echo ""

# 检查Worker是否运行
if ! pgrep -f "seckill-worker" > /dev/null; then
    echo -e "${YELLOW}⚠️  警告: seckill-worker 进程未运行${NC}"
    echo "   建议启动: go run ./cmd/seckill-worker &"
    echo ""
fi

echo "步骤2: 运行功能测试程序..."
echo ""

cd cmd/test-seckill-features
go run main.go

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo ""
echo "提示: 某些测试需要手动验证，请查看测试输出中的说明"

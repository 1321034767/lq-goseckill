#!/bin/bash

echo "=========================================="
echo "GoSeckill 服务状态检查"
echo "=========================================="

echo ""
echo "[1] 检查端口监听状态："
if netstat -tlnp 2>/dev/null | grep -E "8080|8081" > /dev/null; then
    netstat -tlnp | grep -E "8080|8081"
else
    echo "  ❌ 端口 8080 或 8081 未监听"
fi

echo ""
echo "[2] 检查进程状态："
if ps aux | grep -E "go run.*cmd/web|go run.*cmd/admin|bin/web|bin/admin" | grep -v grep > /dev/null; then
    ps aux | grep -E "go run.*cmd/web|go run.*cmd/admin|bin/web|bin/admin" | grep -v grep
else
    echo "  ❌ 服务进程未运行"
fi

echo ""
echo "[3] 检查 MySQL 服务："
if systemctl is-active --quiet mysql || systemctl is-active --quiet mysqld; then
    echo "  ✅ MySQL 服务运行中"
    if mysql -u goseckill -pgoseckill123 -h 127.0.0.1 -e "SELECT 1" > /dev/null 2>&1; then
        echo "  ✅ MySQL 连接正常"
    else
        echo "  ⚠️  MySQL 连接失败，请检查数据库配置"
    fi
else
    echo "  ❌ MySQL 服务未运行"
fi

echo ""
echo "[4] 检查 Redis 服务："
if systemctl is-active --quiet redis-server || systemctl is-active --quiet redis; then
    echo "  ✅ Redis 服务运行中"
    if redis-cli ping > /dev/null 2>&1; then
        echo "  ✅ Redis 连接正常"
    else
        echo "  ⚠️  Redis 连接失败"
    fi
else
    echo "  ❌ Redis 服务未运行"
fi

echo ""
echo "[5] 检查 RabbitMQ 服务："
if systemctl is-active --quiet rabbitmq-server; then
    echo "  ✅ RabbitMQ 服务运行中"
else
    echo "  ❌ RabbitMQ 服务未运行"
fi

echo ""
echo "[6] 测试本地 HTTP 连接："
if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "  ✅ Web 服务 (8080) 本地访问正常"
    curl -s http://localhost:8080/api/health
else
    echo "  ❌ Web 服务 (8080) 本地无法访问"
fi

if curl -s http://localhost:8081/api/health > /dev/null 2>&1; then
    echo "  ✅ Admin 服务 (8081) 本地访问正常"
else
    echo "  ⚠️  Admin 服务 (8081) 本地无法访问（可能未启动）"
fi

echo ""
echo "[7] 检查防火墙："
if command -v ufw > /dev/null; then
    echo "  UFW 状态:"
    ufw status | head -5
elif command -v firewall-cmd > /dev/null; then
    echo "  Firewalld 状态:"
    firewall-cmd --list-ports 2>/dev/null || echo "  未开放端口"
else
    echo "  未检测到防火墙工具"
fi

echo ""
echo "=========================================="
echo "检查完成"
echo "=========================================="
echo ""
echo "如果服务未运行，请执行："
echo "  cd /root/internship-preparation/goseckill-12-01"
echo "  nohup go run ./cmd/web > web.log 2>&1 &"
echo "  nohup go run ./cmd/admin > admin.log 2>&1 &"
echo ""
echo "查看日志："
echo "  tail -f web.log"
echo "  tail -f admin.log"
echo ""

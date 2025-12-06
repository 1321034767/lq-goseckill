#!/bin/bash

# GoSeckill 快速部署脚本
# 使用方法: sudo bash deploy.sh

set -e

PROJECT_DIR="/root/go-seckill/lq-goseckill"
SERVICE_USER="root"

echo "=========================================="
echo "GoSeckill 秒杀系统部署脚本"
echo "=========================================="

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then 
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 1. 检查并安装 Go
echo "[1/8] 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "Go 未安装，正在安装..."
    GO_VERSION="1.22.0"
    ARCH="linux-amd64"
    wget -q https://go.dev/dl/go${GO_VERSION}.${ARCH}.tar.gz
    tar -C /usr/local -xzf go${GO_VERSION}.${ARCH}.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    rm go${GO_VERSION}.${ARCH}.tar.gz
    echo "Go 安装完成"
else
    echo "Go 已安装: $(go version)"
fi

# 2. 检查并安装 MySQL
echo "[2/8] 检查 MySQL..."
if ! command -v mysql &> /dev/null; then
    echo "MySQL 未安装，正在安装..."
    if [ -f /etc/debian_version ]; then
        apt update
        apt install -y mysql-server
    elif [ -f /etc/redhat-release ]; then
        yum install -y mysql-server
        systemctl start mysqld
        systemctl enable mysqld
    fi
    echo "MySQL 安装完成"
else
    echo "MySQL 已安装"
fi

# 启动 MySQL
systemctl start mysql 2>/dev/null || systemctl start mysqld
systemctl enable mysql 2>/dev/null || systemctl enable mysqld

# 创建数据库（如果不存在）
mysql -u root <<EOF 2>/dev/null || echo "请手动创建数据库: CREATE DATABASE goseckill;"
CREATE DATABASE IF NOT EXISTS goseckill CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'goseckill'@'localhost' IDENTIFIED BY 'goseckill123';
GRANT ALL PRIVILEGES ON goseckill.* TO 'goseckill'@'localhost';
FLUSH PRIVILEGES;
EOF

# 3. 检查并安装 Redis
echo "[3/8] 检查 Redis..."
if ! command -v redis-cli &> /dev/null; then
    echo "Redis 未安装，正在安装..."
    if [ -f /etc/debian_version ]; then
        apt install -y redis-server
    elif [ -f /etc/redhat-release ]; then
        yum install -y epel-release
        yum install -y redis
    fi
    echo "Redis 安装完成"
else
    echo "Redis 已安装"
fi

# 启动 Redis
systemctl start redis-server 2>/dev/null || systemctl start redis
systemctl enable redis-server 2>/dev/null || systemctl enable redis

# 4. 检查并安装 RabbitMQ
echo "[4/8] 检查 RabbitMQ..."
if ! command -v rabbitmqctl &> /dev/null; then
    echo "RabbitMQ 未安装，正在安装..."
    if [ -f /etc/debian_version ]; then
        apt install -y rabbitmq-server
    elif [ -f /etc/redhat-release ]; then
        yum install -y rabbitmq-server
    fi
    echo "RabbitMQ 安装完成"
else
    echo "RabbitMQ 已安装"
fi

# 启动 RabbitMQ
systemctl start rabbitmq-server
systemctl enable rabbitmq-server

# 5. 检查项目目录
echo "[5/8] 检查项目目录..."
if [ ! -d "$PROJECT_DIR" ]; then
    echo "错误: 项目目录不存在: $PROJECT_DIR"
    echo "请先将项目文件上传到 $PROJECT_DIR"
    exit 1
fi

cd "$PROJECT_DIR"

# 6. 安装 Go 依赖
echo "[6/8] 安装 Go 依赖..."
export PATH=$PATH:/usr/local/go/bin
go mod download
go mod tidy

# 7. 编译项目
echo "[7/8] 编译项目..."
mkdir -p bin
go build -o bin/web ./cmd/web
go build -o bin/admin ./cmd/admin
go build -o bin/seckill-worker ./cmd/seckill-worker
echo "编译完成"

# 8. 创建 systemd 服务
echo "[8/8] 创建 systemd 服务..."

# Web 服务
cat > /etc/systemd/system/goseckill-web.service <<EOF
[Unit]
Description=GoSeckill Web Service
After=network.target mysql.service redis-server.service rabbitmq-server.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$PROJECT_DIR
ExecStart=$PROJECT_DIR/bin/web
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Admin 服务
cat > /etc/systemd/system/goseckill-admin.service <<EOF
[Unit]
Description=GoSeckill Admin Service
After=network.target mysql.service redis-server.service rabbitmq-server.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$PROJECT_DIR
ExecStart=$PROJECT_DIR/bin/admin
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Worker 服务
cat > /etc/systemd/system/goseckill-worker.service <<EOF
[Unit]
Description=GoSeckill Worker Service
After=network.target mysql.service redis-server.service rabbitmq-server.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$PROJECT_DIR
ExecStart=$PROJECT_DIR/bin/seckill-worker
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd
systemctl daemon-reload

# 启动服务
echo "启动服务..."
systemctl enable goseckill-web
systemctl enable goseckill-admin
systemctl enable goseckill-worker

systemctl start goseckill-web
systemctl start goseckill-admin
systemctl start goseckill-worker

# 等待服务启动
sleep 3

# 检查服务状态
echo ""
echo "=========================================="
echo "部署完成！"
echo "=========================================="
echo ""
echo "服务状态："
systemctl status goseckill-web --no-pager -l | head -5
systemctl status goseckill-admin --no-pager -l | head -5
systemctl status goseckill-worker --no-pager -l | head -5
echo ""
echo "访问地址："
echo "  - Web 前端: http://$(hostname -I | awk '{print $1}'):8080"
echo "  - Admin 后台: http://$(hostname -I | awk '{print $1}'):8081"
echo ""
echo "常用命令："
echo "  查看日志: sudo journalctl -u goseckill-web -f"
echo "  重启服务: sudo systemctl restart goseckill-web"
echo "  停止服务: sudo systemctl stop goseckill-web"
echo ""
echo "详细部署说明请查看: $PROJECT_DIR/部署指南.md"
echo "=========================================="

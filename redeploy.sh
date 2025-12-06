#!/bin/bash

# GoSeckill 快速重新部署脚本
# 用于代码更新后重新编译和重启服务

set -e

PROJECT_DIR="${1:-/root/go-seckill/lq-goseckill}"

ok()   { echo -e "  ✅ $*"; }
warn() { echo -e "  ⚠️  $*"; }
err()  { echo -e "  ❌ $*"; }
section() { echo -e "\n==================== $* ===================="; }

echo "=========================================="
echo "GoSeckill 快速重新部署"
echo "=========================================="
echo "项目目录: $PROJECT_DIR"
echo ""

# 检查项目目录
if [ ! -d "$PROJECT_DIR" ]; then
    err "项目目录不存在: $PROJECT_DIR"
    exit 1
fi

cd "$PROJECT_DIR"

# 1. 检查 Go 环境
section "1. 检查 Go 环境"
if ! command -v go &> /dev/null; then
    err "Go 未安装，请先安装 Go"
    exit 1
else
    ok "Go 版本: $(go version)"
fi

# 2. 更新依赖
section "2. 更新 Go 依赖"
export PATH=$PATH:/usr/local/go/bin
go mod download
go mod tidy
ok "依赖更新完成"

# 3. 停止现有服务
section "3. 停止现有服务"

if systemctl is-active --quiet goseckill-web 2>/dev/null; then
    echo "  正在停止 goseckill-web..."
    sudo systemctl stop goseckill-web
    ok "goseckill-web 已停止"
else
    warn "goseckill-web 未运行"
fi

if systemctl is-active --quiet goseckill-admin 2>/dev/null; then
    echo "  正在停止 goseckill-admin..."
    sudo systemctl stop goseckill-admin
    ok "goseckill-admin 已停止"
else
    warn "goseckill-admin 未运行"
fi

if systemctl is-active --quiet goseckill-worker 2>/dev/null; then
    echo "  正在停止 goseckill-worker..."
    sudo systemctl stop goseckill-worker
    ok "goseckill-worker 已停止"
else
    warn "goseckill-worker 未运行"
fi

# 检查是否有手动运行的进程
if pgrep -f "go run.*cmd/web" > /dev/null; then
    warn "发现手动运行的 web 进程，正在停止..."
    pkill -f "go run.*cmd/web" || true
fi

if pgrep -f "go run.*cmd/admin" > /dev/null; then
    warn "发现手动运行的 admin 进程，正在停止..."
    pkill -f "go run.*cmd/admin" || true
fi

# 4. 编译项目
section "4. 编译项目"
mkdir -p bin

echo "  正在编译 web 服务..."
if go build -o bin/web ./cmd/web; then
    ok "web 编译成功"
else
    err "web 编译失败"
    exit 1
fi

echo "  正在编译 admin 服务..."
if go build -o bin/admin ./cmd/admin; then
    ok "admin 编译成功"
else
    err "admin 编译失败"
    exit 1
fi

echo "  正在编译 seckill-worker..."
if go build -o bin/seckill-worker ./cmd/seckill-worker; then
    ok "seckill-worker 编译成功"
else
    warn "seckill-worker 编译失败（可能不是必需的）"
fi

# 5. 检查编译产物
section "5. 检查编译产物"
if [ -f "bin/web" ]; then
    ok "bin/web 存在 ($(du -h bin/web | cut -f1))"
else
    err "bin/web 不存在"
    exit 1
fi

if [ -f "bin/admin" ]; then
    ok "bin/admin 存在 ($(du -h bin/admin | cut -f1))"
else
    err "bin/admin 不存在"
    exit 1
fi

# 6. 启动服务
section "6. 启动服务"

# 检查是否使用 systemd
if [ -f "/etc/systemd/system/goseckill-web.service" ]; then
    echo "  使用 systemd 管理服务..."
    
    sudo systemctl daemon-reload
    
    echo "  启动 goseckill-web..."
    sudo systemctl start goseckill-web
    sleep 2
    
    echo "  启动 goseckill-admin..."
    sudo systemctl start goseckill-admin
    sleep 2
    
    if [ -f "/etc/systemd/system/goseckill-worker.service" ]; then
        echo "  启动 goseckill-worker..."
        sudo systemctl start goseckill-worker
        sleep 2
    fi
    
    # 检查服务状态
    echo ""
    echo "  服务状态："
    systemctl is-active --quiet goseckill-web && ok "goseckill-web: 运行中" || err "goseckill-web: 未运行"
    systemctl is-active --quiet goseckill-admin && ok "goseckill-admin: 运行中" || err "goseckill-admin: 未运行"
    systemctl is-active --quiet goseckill-worker && ok "goseckill-worker: 运行中" || warn "goseckill-worker: 未运行"
    
else
    echo "  未找到 systemd 服务，使用后台运行方式..."
    
    # 创建日志目录
    mkdir -p logs
    
    echo "  启动 web 服务 (8080)..."
    nohup ./bin/web > logs/web.log 2>&1 &
    WEB_PID=$!
    sleep 2
    
    echo "  启动 admin 服务 (8081)..."
    nohup ./bin/admin > logs/admin.log 2>&1 &
    ADMIN_PID=$!
    sleep 2
    
    if [ -f "bin/seckill-worker" ]; then
        echo "  启动 seckill-worker..."
        nohup ./bin/seckill-worker > logs/worker.log 2>&1 &
        WORKER_PID=$!
        sleep 2
    fi
    
    # 检查进程
    echo ""
    echo "  进程状态："
    if ps -p $WEB_PID > /dev/null 2>&1; then
        ok "web 进程运行中 (PID: $WEB_PID)"
    else
        err "web 进程启动失败，查看日志: tail -f logs/web.log"
    fi
    
    if ps -p $ADMIN_PID > /dev/null 2>&1; then
        ok "admin 进程运行中 (PID: $ADMIN_PID)"
    else
        err "admin 进程启动失败，查看日志: tail -f logs/admin.log"
    fi
fi

# 7. 检查端口监听
section "7. 检查端口监听"
sleep 3

if command -v ss >/dev/null 2>&1; then
    if ss -tlnp | grep -q ":8080 "; then
        ok "端口 8080 正在监听"
    else
        err "端口 8080 未监听"
    fi
    
    if ss -tlnp | grep -q ":8081 "; then
        ok "端口 8081 正在监听"
    else
        err "端口 8081 未监听"
    fi
fi

# 8. 测试服务
section "8. 测试服务"

if command -v curl >/dev/null 2>&1; then
    echo "  测试 web 服务..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health | grep -q "200\|404"; then
        ok "web 服务响应正常"
    else
        warn "web 服务可能未就绪，稍等片刻后重试"
    fi
    
    echo "  测试 admin 服务..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/api/health | grep -q "200"; then
        ok "admin 服务响应正常"
    else
        warn "admin 服务可能未就绪，稍等片刻后重试"
    fi
fi

# 完成
echo ""
echo "=========================================="
echo "重新部署完成！"
echo "=========================================="
echo ""
echo "访问地址："
echo "  - Web 前台: http://akaina.site 或 http://你的服务器IP:8080"
echo "  - Admin 后台: http://admin.akaina.site 或 http://你的服务器IP:8081"
echo ""
echo "常用命令："
if [ -f "/etc/systemd/system/goseckill-web.service" ]; then
    echo "  查看日志: sudo journalctl -u goseckill-web -f"
    echo "  重启服务: sudo systemctl restart goseckill-web"
    echo "  查看状态: sudo systemctl status goseckill-web"
else
    echo "  查看日志: tail -f logs/web.log"
    echo "  查看日志: tail -f logs/admin.log"
    echo "  停止服务: pkill -f bin/web"
fi
echo ""
echo "如果服务未正常启动，请检查："
echo "  1. MySQL 是否运行: sudo systemctl status mysql"
echo "  2. Redis 是否运行: sudo systemctl status redis-server"
echo "  3. RabbitMQ 是否运行: sudo systemctl status rabbitmq-server"
echo "  4. 查看服务日志获取详细错误信息"
echo ""

#!/bin/bash

echo "=========================================="
echo "配置防火墙开放端口"
echo "=========================================="

# 检查 UFW 是否启用
if command -v ufw > /dev/null; then
    echo "[1] 检测到 UFW 防火墙"
    echo "当前状态："
    ufw status
    
    echo ""
    echo "[2] 开放端口 8080 和 8081..."
    sudo ufw allow 8080/tcp
    sudo ufw allow 8081/tcp
    
    echo ""
    echo "[3] 重新加载防火墙规则..."
    sudo ufw reload
    
    echo ""
    echo "[4] 检查规则是否生效："
    ufw status | grep -E "8080|8081"
    
    echo ""
    echo "✅ 防火墙配置完成！"
    
elif command -v firewall-cmd > /dev/null; then
    echo "[1] 检测到 Firewalld 防火墙"
    echo "当前状态："
    firewall-cmd --list-ports
    
    echo ""
    echo "[2] 开放端口 8080 和 8081..."
    sudo firewall-cmd --permanent --add-port=8080/tcp
    sudo firewall-cmd --permanent --add-port=8081/tcp
    
    echo ""
    echo "[3] 重新加载防火墙规则..."
    sudo firewall-cmd --reload
    
    echo ""
    echo "[4] 检查规则是否生效："
    firewall-cmd --list-ports | grep -E "8080|8081"
    
    echo ""
    echo "✅ 防火墙配置完成！"
    
else
    echo "⚠️  未检测到防火墙工具，可能需要手动配置"
    echo "或者检查云服务器的安全组设置"
fi

echo ""
echo "=========================================="
echo "重要提示"
echo "=========================================="
echo "1. 如果这是云服务器，还需要在控制台配置安全组规则"
echo "2. 开放端口：8080 (Web) 和 8081 (Admin)"
echo "3. 授权对象：0.0.0.0/0（允许所有IP）或指定你的IP"
echo ""
echo "配置完成后，访问："
echo "  Web: http://$(hostname -I | awk '{print $1}'):8080"
echo "  Admin: http://$(hostname -I | awk '{print $1}'):8081"
echo "=========================================="

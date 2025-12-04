#!/bin/bash

echo "=========================================="
echo "检查并添加商品数据"
echo "=========================================="

# 检查商品详情页
echo ""
echo "1. 检查商品详情页 /product/1..."
response=$(curl -s -o /tmp/check_product.html -w "%{http_code}" http://localhost:8080/product/1)
echo "HTTP状态码: $response"

if grep -q "商品不存在\|异常错误处理页面" /tmp/check_product.html 2>/dev/null; then
    echo "❌ 商品不存在"
    echo ""
    echo "2. 需要添加商品数据"
    echo ""
    echo "有两种方式添加商品："
    echo ""
    echo "方式1: 通过后台管理API添加（推荐）"
    echo "  步骤1: 启动后台管理服务器（如果未运行）"
    echo "    cd /root/internship-preparation/goseckill-12-01"
    echo "    go run ./cmd/admin"
    echo ""
    echo "  步骤2: 运行批量添加商品脚本"
    echo "    go run ./cmd/add-products/main.go"
    echo ""
    echo "方式2: 直接通过数据库添加（需要数据库访问权限）"
    echo ""
    echo "=========================================="
    exit 1
else
    echo "✅ 商品存在，页面应该能正常显示"
    echo ""
    echo "如果页面仍然没有内容，请检查："
    echo "1. 浏览器控制台是否有JavaScript错误"
    echo "2. CSS/JS文件是否加载成功"
    echo "3. 查看完整HTML: cat /tmp/check_product.html"
fi

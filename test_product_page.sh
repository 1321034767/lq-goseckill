#!/bin/bash

# 测试商品详情页是否能正常显示

echo "=========================================="
echo "商品详情页测试脚本"
echo "=========================================="

# 检查服务器是否运行
echo "1. 检查服务器是否运行..."
if ! curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "❌ 错误: 服务器未运行，请先启动服务器"
    echo "   启动命令: cd /root/internship-preparation/goseckill-12-01 && go run ./cmd/web"
    exit 1
fi
echo "✅ 服务器正在运行"

# 测试商品详情页
echo ""
echo "2. 测试商品详情页 /product/1..."
response=$(curl -s -o /tmp/product_page.html -w "%{http_code}" http://localhost:8080/product/1)

if [ "$response" != "200" ]; then
    echo "❌ 错误: HTTP状态码 $response (期望 200)"
    exit 1
fi
echo "✅ HTTP状态码: $response"

# 检查HTML内容
echo ""
echo "3. 检查HTML内容..."

# 检查是否有loader-mask
if grep -q "loader-mask" /tmp/product_page.html; then
    echo "✅ 找到 loader-mask 元素"
else
    echo "⚠️  警告: 未找到 loader-mask 元素"
fi

# 检查是否有隐藏loader的脚本
if grep -q "hideLoader\|loader-mask.*style.*display.*none\|loader-mask.*remove" /tmp/product_page.html; then
    echo "✅ 找到隐藏loader的脚本"
else
    echo "❌ 错误: 未找到隐藏loader的脚本"
fi

# 检查是否有商品内容
if grep -q "product-single\|商品详情" /tmp/product_page.html; then
    echo "✅ 找到商品详情内容"
else
    echo "❌ 错误: 未找到商品详情内容"
fi

# 检查是否有导航栏
if grep -q "nav.*header\|top-bar" /tmp/product_page.html; then
    echo "✅ 找到导航栏"
else
    echo "❌ 错误: 未找到导航栏"
fi

# 检查是否有JavaScript文件引用
if grep -q "scripts.js\|jquery.min.js" /tmp/product_page.html; then
    echo "✅ 找到JavaScript文件引用"
else
    echo "❌ 错误: 未找到JavaScript文件引用"
fi

# 检查HTML结构
echo ""
echo "4. 检查HTML结构..."
if grep -q "<!DOCTYPE html>" /tmp/product_page.html; then
    echo "✅ HTML文档类型正确"
else
    echo "❌ 错误: HTML文档类型不正确"
fi

if grep -q "<html" /tmp/product_page.html && grep -q "</html>" /tmp/product_page.html; then
    echo "✅ HTML结构完整"
else
    echo "❌ 错误: HTML结构不完整"
fi

# 检查是否有JavaScript错误（通过检查关键函数）
echo ""
echo "5. 检查关键JavaScript函数..."
if grep -q "function hideLoader\|hideLoader()" /tmp/product_page.html; then
    echo "✅ 找到 hideLoader 函数"
else
    echo "❌ 错误: 未找到 hideLoader 函数"
fi

if grep -q "getCookie" /tmp/product_page.html; then
    echo "✅ 找到 getCookie 函数"
else
    echo "⚠️  警告: 未找到 getCookie 函数"
fi

# 显示HTML文件大小
echo ""
echo "6. HTML文件信息:"
file_size=$(wc -c < /tmp/product_page.html)
echo "   文件大小: $file_size 字节"
line_count=$(wc -l < /tmp/product_page.html)
echo "   行数: $line_count"

# 显示前50行（包含head和body开始部分）
echo ""
echo "7. HTML文件前50行预览:"
head -50 /tmp/product_page.html

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="
echo ""
echo "如果页面仍然显示加载动画，请检查："
echo "1. 浏览器控制台是否有JavaScript错误"
echo "2. 网络面板中CSS/JS文件是否加载成功"
echo "3. 检查 /tmp/product_page.html 文件查看完整HTML"
echo ""

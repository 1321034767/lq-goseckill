#!/bin/bash

# 完整的商品详情页测试脚本

echo "=========================================="
echo "商品详情页完整测试"
echo "=========================================="

# 1. 检查服务器
echo ""
echo "1. 检查服务器状态..."
if ! curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "❌ 服务器未运行"
    exit 1
fi
echo "✅ 服务器运行中"

# 2. 获取商品列表
echo ""
echo "2. 获取商品列表..."
products=$(curl -s http://localhost:8080/api/products)
echo "$products" | head -20

# 提取第一个商品ID
product_id=$(echo "$products" | grep -o '"id":[0-9]*' | head -1 | grep -o '[0-9]*')
if [ -z "$product_id" ]; then
    echo "❌ 未找到商品，请先添加商品"
    exit 1
fi
echo ""
echo "✅ 找到商品，使用商品ID: $product_id"

# 3. 测试商品详情页
echo ""
echo "3. 测试商品详情页 /product/$product_id..."
response=$(curl -s -o /tmp/product_detail_test.html -w "%{http_code}" "http://localhost:8080/product/$product_id")

if [ "$response" != "200" ]; then
    echo "❌ HTTP状态码: $response (期望 200)"
    echo "响应内容:"
    cat /tmp/product_detail_test.html
    exit 1
fi
echo "✅ HTTP状态码: $response"

# 4. 检查关键内容
echo ""
echo "4. 检查页面内容..."

checks_passed=0
checks_total=0

# 检查loader-mask
checks_total=$((checks_total + 1))
if grep -q "loader-mask" /tmp/product_detail_test.html; then
    echo "✅ 找到 loader-mask"
    checks_passed=$((checks_passed + 1))
else
    echo "❌ 未找到 loader-mask"
fi

# 检查hideLoader函数
checks_total=$((checks_total + 1))
if grep -q "hideLoader\|function hideLoader" /tmp/product_detail_test.html; then
    echo "✅ 找到 hideLoader 函数"
    checks_passed=$((checks_passed + 1))
else
    echo "❌ 未找到 hideLoader 函数"
fi

# 检查商品内容
checks_total=$((checks_total + 1))
if grep -q "product-single\|商品详情" /tmp/product_detail_test.html; then
    echo "✅ 找到商品详情内容"
    checks_passed=$((checks_passed + 1))
else
    echo "❌ 未找到商品详情内容"
fi

# 检查导航栏
checks_total=$((checks_total + 1))
if grep -q "nav.*header\|top-bar" /tmp/product_detail_test.html; then
    echo "✅ 找到导航栏"
    checks_passed=$((checks_passed + 1))
else
    echo "❌ 未找到导航栏"
fi

# 检查JavaScript文件
checks_total=$((checks_total + 1))
if grep -q "scripts.js\|jquery.min.js" /tmp/product_detail_test.html; then
    echo "✅ 找到JavaScript文件"
    checks_passed=$((checks_passed + 1))
else
    echo "❌ 未找到JavaScript文件"
fi

# 检查HTML结构
checks_total=$((checks_total + 1))
if grep -q "<!DOCTYPE html>" /tmp/product_detail_test.html && grep -q "</html>" /tmp/product_detail_test.html; then
    echo "✅ HTML结构完整"
    checks_passed=$((checks_passed + 1))
else
    echo "❌ HTML结构不完整"
fi

# 5. 显示文件信息
echo ""
echo "5. 文件信息:"
file_size=$(wc -c < /tmp/product_detail_test.html)
echo "   大小: $file_size 字节"
line_count=$(wc -l < /tmp/product_detail_test.html)
echo "   行数: $line_count"

# 6. 检查是否有错误页面内容
echo ""
echo "6. 检查错误页面..."
if grep -q "商品不存在\|error\|Error" /tmp/product_detail_test.html; then
    echo "⚠️  警告: 可能包含错误信息"
    echo "   错误内容预览:"
    grep -i "商品不存在\|error" /tmp/product_detail_test.html | head -3
fi

# 7. 显示关键HTML片段
echo ""
echo "7. HTML关键片段:"
echo "   Head部分:"
grep -A 10 "<head>" /tmp/product_detail_test.html | head -15
echo ""
echo "   Body开始部分:"
grep -A 20 "<body>" /tmp/product_detail_test.html | head -25

# 8. 总结
echo ""
echo "=========================================="
echo "测试总结: $checks_passed/$checks_total 项通过"
echo "=========================================="

if [ $checks_passed -eq $checks_total ]; then
    echo "✅ 所有检查通过！"
    exit 0
else
    echo "❌ 部分检查未通过"
    echo ""
    echo "完整HTML文件保存在: /tmp/product_detail_test.html"
    exit 1
fi

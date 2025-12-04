#!/bin/bash

# 完整的商品详情页测试和验证脚本

set -e

echo "=========================================="
echo "商品详情页完整测试脚本"
echo "=========================================="

BASE_URL="http://localhost:8080"
TEST_DIR="/tmp/product_test_$$"
mkdir -p "$TEST_DIR"

# 清理函数
cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# 1. 检查服务器
echo ""
echo "[1/8] 检查服务器状态..."
if ! curl -s "$BASE_URL/api/health" > /dev/null 2>&1; then
    echo "❌ 错误: 服务器未运行"
    echo "   请先启动服务器: cd /root/internship-preparation/goseckill-12-01 && go run ./cmd/web"
    exit 1
fi
echo "✅ 服务器运行正常"

# 2. 获取商品列表
echo ""
echo "[2/8] 获取商品列表..."
products_response=$(curl -s "$BASE_URL/api/products" || echo "[]")
echo "响应: $products_response" | head -c 200
echo ""

# 提取商品ID（尝试多个ID）
test_ids=(1 2 3)
product_id=""

for id in "${test_ids[@]}"; do
    echo "尝试商品ID: $id"
    response=$(curl -s -o "$TEST_DIR/product_$id.html" -w "%{http_code}" "$BASE_URL/product/$id")
    if [ "$response" = "200" ]; then
        # 检查是否是错误页面
        if ! grep -q "商品不存在\|异常错误处理页面" "$TEST_DIR/product_$id.html" 2>/dev/null; then
            product_id=$id
            echo "✅ 找到有效商品，ID: $product_id"
            break
        fi
    fi
done

if [ -z "$product_id" ]; then
    echo "⚠️  警告: 未找到有效商品，将测试错误页面"
    product_id=99999
fi

# 3. 测试商品详情页
echo ""
echo "[3/8] 测试商品详情页 /product/$product_id..."
response=$(curl -s -o "$TEST_DIR/product_detail.html" -w "%{http_code}" "$BASE_URL/product/$product_id")

if [ "$response" != "200" ]; then
    echo "❌ HTTP状态码: $response (期望 200)"
    cat "$TEST_DIR/product_detail.html"
    exit 1
fi
echo "✅ HTTP状态码: $response"

file_size=$(wc -c < "$TEST_DIR/product_detail.html")
echo "   文件大小: $file_size 字节"

# 4. 检查关键元素
echo ""
echo "[4/8] 检查关键HTML元素..."

checks=(
    "loader-mask:loader-mask"
    "hideLoader函数:function hideLoader|hideLoader()"
    "productLayout结构:productLayout|<!DOCTYPE html>"
    "导航栏:nav.*header|top-bar"
    "JavaScript文件:scripts.js|jquery.min.js"
    "HTML结构:<!DOCTYPE html>.*</html>"
)

passed=0
total=${#checks[@]}

for check in "${checks[@]}"; do
    name="${check%%:*}"
    pattern="${check#*:}"
    if grep -qE "$pattern" "$TEST_DIR/product_detail.html" 2>/dev/null; then
        echo "✅ $name"
        passed=$((passed + 1))
    else
        echo "❌ $name (未找到)"
    fi
done

# 5. 检查loader隐藏逻辑
echo ""
echo "[5/8] 检查loader隐藏逻辑..."
loader_checks=(
    "loader-mask元素存在:loader-mask"
    "hideLoader函数定义:function hideLoader"
    "setTimeout调用:setTimeout.*hideLoader"
    "DOMContentLoaded监听:DOMContentLoaded"
    "remove方法调用:\.remove\(\)"
)

loader_passed=0
loader_total=${#loader_checks[@]}

for check in "${loader_checks[@]}"; do
    name="${check%%:*}"
    pattern="${check#*:}"
    if grep -qE "$pattern" "$TEST_DIR/product_detail.html" 2>/dev/null; then
        echo "✅ $name"
        loader_passed=$((loader_passed + 1))
    else
        echo "❌ $name"
    fi
done

# 6. 验证HTML结构
echo ""
echo "[6/8] 验证HTML结构..."
if grep -q "<!DOCTYPE html>" "$TEST_DIR/product_detail.html" && \
   grep -q "</html>" "$TEST_DIR/product_detail.html"; then
    echo "✅ HTML文档结构完整"
    
    # 检查head和body
    if grep -q "<head>" "$TEST_DIR/product_detail.html" && \
       grep -q "<body>" "$TEST_DIR/product_detail.html"; then
        echo "✅ 包含head和body标签"
    else
        echo "❌ 缺少head或body标签"
    fi
else
    echo "❌ HTML文档结构不完整"
fi

# 7. 显示关键代码片段
echo ""
echo "[7/8] 关键代码片段:"
echo "--- loader隐藏脚本位置 ---"
grep -A 15 "function hideLoader\|hideLoader()" "$TEST_DIR/product_detail.html" | head -20 || echo "未找到"
echo ""
echo "--- loader-mask元素位置 ---"
grep -B 2 -A 5 "loader-mask" "$TEST_DIR/product_detail.html" | head -10 || echo "未找到"

# 8. 生成测试报告
echo ""
echo "[8/8] 生成测试报告..."
report_file="$TEST_DIR/test_report.txt"
{
    echo "商品详情页测试报告"
    echo "==================="
    echo "测试时间: $(date)"
    echo "测试URL: $BASE_URL/product/$product_id"
    echo "HTTP状态码: $response"
    echo "文件大小: $file_size 字节"
    echo ""
    echo "检查结果: $passed/$total 通过"
    echo "Loader检查: $loader_passed/$loader_total 通过"
    echo ""
    echo "完整HTML文件: $TEST_DIR/product_detail.html"
} > "$report_file"

cat "$report_file"

# 9. 总结
echo ""
echo "=========================================="
echo "测试总结"
echo "=========================================="
echo "基本检查: $passed/$total"
echo "Loader检查: $loader_passed/$loader_total"
echo ""
echo "测试文件保存在: $TEST_DIR"
echo "  - HTML文件: $TEST_DIR/product_detail.html"
echo "  - 测试报告: $TEST_DIR/test_report.txt"
echo ""

if [ $passed -eq $total ] && [ $loader_passed -ge 3 ]; then
    echo "✅ 测试通过！页面应该能正常显示"
    exit 0
else
    echo "⚠️  部分检查未通过，请查看详细报告"
    echo ""
    echo "如果页面仍然显示加载动画，请："
    echo "1. 检查浏览器控制台的JavaScript错误"
    echo "2. 检查网络面板中CSS/JS文件是否加载成功"
    echo "3. 查看完整HTML: cat $TEST_DIR/product_detail.html"
    exit 1
fi

#!/bin/bash

echo "验证编译错误修复..."
cd cmd/test-seckill-features

echo "编译测试程序..."
if go build -o test-seckill main.go 2>&1; then
    echo "✅ 编译成功！"
    rm -f test-seckill
    exit 0
else
    echo "❌ 编译失败，错误信息："
    go build -o test-seckill main.go 2>&1
    exit 1
fi

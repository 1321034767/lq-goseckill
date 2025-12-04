# 商品ID映射说明

## 当前配置

已修改代码，使商品ID 15-26 映射到图片 1-12：

- **商品ID 15** → 图片 `product_1.jpg` / `product_back_1.jpg`
- **商品ID 16** → 图片 `product_2.jpg` / `product_back_2.jpg`
- **商品ID 17** → 图片 `product_3.jpg` / `product_back_3.jpg`
- ...
- **商品ID 26** → 图片 `product_12.jpg` / `product_back_12.jpg`

## 修改的文件

1. **`internal/server/router.go`**
   - 修改了 `getProductImages` 函数中的图片索引计算逻辑
   - 商品ID 3-14 直接映射到图片 1-12
   - 其他ID使用循环映射（1-12循环）

2. **`web/index.html`**
   - 修改了首页商品列表的图片索引计算逻辑
   - 与后端保持一致

## 如果需要将数据库中的商品ID改为1-12

如果您希望将数据库中的商品ID从 3-14 改为 1-12，可以使用以下方法：

### 方法1：删除并重新添加（推荐）

1. **删除所有现有商品**：
   ```bash
   go run ./cmd/reset-products-to-1-12/main.go
   ```

2. **重新添加12个商品**（ID将自动从1开始）：
   ```bash
   go run ./cmd/add-products/main.go
   ```

这样新添加的商品ID将是：1, 2, 3, ..., 12

### 方法2：手动删除特定ID的商品

如果您只想删除ID为1和2的商品（如果存在），可以使用：
```bash
go run ./cmd/delete-test-products/main.go
```

## 注意事项

- 如果删除商品，相关的订单数据可能会受到影响（如果订单表中有外键约束）
- 建议在删除前备份数据库
- 图片索引计算已经修改，无论商品ID是什么，都会正确映射到图片1-12

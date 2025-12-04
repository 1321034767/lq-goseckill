# 商品图片文件夹结构

## 目录结构

商品图片现在按商品ID组织在独立的文件夹中：

```
web/assets/img/shop/
├── 1/
│   ├── product_1.jpg
│   └── product_back_1.jpg
├── 2/
│   ├── product_2.jpg
│   └── product_back_2.jpg
├── 3/
│   ├── product_3.jpg
│   └── product_back_3.jpg
...
└── 12/
    ├── product_12.jpg
    └── product_back_12.jpg
```

## 图片路径

### 商品详情页
- 主图：`/assets/img/shop/{商品ID}/product_{商品ID}.jpg`
- 背面图：`/assets/img/shop/{商品ID}/product_back_{商品ID}.jpg`

### 示例
- 商品ID 1：
  - `/assets/img/shop/1/product_1.jpg`
  - `/assets/img/shop/1/product_back_1.jpg`
- 商品ID 5：
  - `/assets/img/shop/5/product_5.jpg`
  - `/assets/img/shop/5/product_back_5.jpg`

## 代码更新

以下文件已更新以使用新的文件夹结构：

1. **后端** (`internal/server/router.go`)
   - `getProductImages` 函数已更新路径

2. **前端** (`web/index.html`)
   - 商品列表的图片路径已更新

## 优势

- ✅ 更好的组织结构：每个商品的图片独立存放
- ✅ 易于管理：可以轻松为每个商品添加更多图片
- ✅ 清晰的映射关系：文件夹编号 = 商品ID

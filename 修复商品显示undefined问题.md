# 修复商品显示undefined问题

## 问题描述

首页商品列表显示 "undefined" 和 "$NaN"，原因是API返回的字段名是小写（`id`、`name`、`price`），而前端JavaScript代码期望的是大写（`ID`、`Name`、`Price`）。

## 问题原因

在修改商品列表API时，返回的数据结构使用了小写字段名：
```go
productData := map[string]interface{}{
    "id":    p.ID,
    "name":  p.Name,
    "price": p.Price,
    // ...
}
```

但前端代码期望的是：
```javascript
p.ID    // 而不是 p.id
p.Name  // 而不是 p.name
p.Price // 而不是 p.price
```

## 解决方案

修改 `internal/server/router.go` 中的商品列表API，将返回的字段名改为大写，以匹配前端期望：

```go
productData := map[string]interface{}{
    "ID":            p.ID,  // 大写ID
    "Name":          p.Name, // 大写Name
    "Price":         p.Price, // 大写Price
    "Stock":         p.Stock,
    "SeckillStock":  p.SeckillStock,
    "Status":        p.Status,
    "StartTime":     p.StartTime,
    "EndTime":       p.EndTime,
    "Description":   p.Description,
    "Category":      p.Category,
}
```

## 修改的文件

- `internal/server/router.go`: 修改商品列表API返回的字段名为大写

## 测试方法

1. 重启web服务
2. 刷新首页
3. 确认商品名称、价格正常显示，不再显示 "undefined" 和 "$NaN"

## 注意事项

- 保持API返回字段名与前端期望一致
- 如果后续修改前端代码，需要同步更新字段名
- 建议统一使用大写字段名（Go的JSON序列化默认使用字段名，但手动构建map时需要保持一致）

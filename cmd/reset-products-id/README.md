# 重置商品ID为1-12

## 功能说明

这个工具会：
1. 通过API删除所有现有商品
2. **重置数据库AUTO_INCREMENT为1**（关键步骤）
3. 重新添加12个商品（ID将从1开始）
4. 验证商品ID范围

## 使用方法

```bash
go run ./cmd/reset-products-id/main.go
```

## 前提条件

- 管理后台服务必须运行在 `localhost:8081`
- 数据库连接配置正确（在 `internal/config/config.go` 中）

## 验证

运行测试程序验证商品ID：

```bash
go run ./cmd/test-product-ids/main.go
```

## 注意事项

- 此操作会**删除所有现有商品**
- 如果有关联的订单数据，可能会受到影响
- 建议在操作前备份数据库

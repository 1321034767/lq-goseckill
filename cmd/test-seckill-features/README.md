# 秒杀功能测试程序

## 功能说明

此测试程序用于验证秒杀系统的12个功能点是否成功实现。

## 测试内容

### 高优先级测试（5项）

1. **秒杀时间校验** - 验证是否在秒杀时间外被拒绝
2. **商品状态校验** - 验证非秒杀商品是否被拒绝
3. **库存自动同步机制** - 需要管理员权限测试
4. **RabbitMQ消息确认机制** - 需要检查Worker日志
5. **Worker失败时的库存回滚** - 需要模拟失败场景

### 中优先级测试（3项）

6. **订单查询接口实现** - 测试 `/api/orders` 接口
7. **秒杀结果查询接口** - 测试 `/api/seckill/{id}/result` 接口
8. **前端秒杀结果展示** - 需要浏览器测试

### 低优先级测试（4项）

9. **秒杀库存实时显示** - 测试 `/api/products/{id}/seckill-stock` 接口
10. **监控功能** - 测试 `/api/monitor/stats` 接口
11. **限流功能** - 测试快速请求是否被限流
12. **库存一致性检查** - 需要运行stock-sync服务

## 运行前准备

### 1. 确保服务已启动

```bash
# 启动依赖服务
sudo systemctl start mysql
sudo systemctl start redis-server
sudo systemctl start rabbitmq-server

# 启动应用服务
go run ./cmd/web &
go run ./cmd/admin &
go run ./cmd/seckill-worker &
```

或者使用部署脚本：

```bash
bash deploy.sh
```

### 2. 确保有测试商品

确保至少有一个商品存在，建议ID为1的商品设置为秒杀状态（Status=2）。

## 运行测试

```bash
cd cmd/test-seckill-features
go run main.go
```

## 预期输出

```
==========================================
    秒杀功能测试程序
==========================================

【准备阶段】注册/登录用户...
✅ 登录成功，Token: eyJhbGciOiJIUzI1NiIs...

==========================================
    开始功能测试
==========================================

测试: 1. 秒杀时间校验
✅ 1. 秒杀时间校验
   结果: 时间校验生效: seckill not started yet
   耗时: 123ms

测试: 2. 商品状态校验
✅ 2. 商品状态校验
   结果: 状态校验生效: product is not in seckill status
   耗时: 98ms

...

==========================================
    测试报告
==========================================

✅ 1. 秒杀时间校验
   结果: 时间校验生效: seckill not started yet
   耗时: 123ms

...

==========================================
总计: 12 个测试
通过: 10 个
失败: 2 个
通过率: 83.3%
==========================================
```

## 手动测试说明

某些测试需要手动验证：

### 1. 库存自动同步机制

```bash
# 1. 通过Admin接口更新商品库存
curl -X PUT http://localhost:8081/api/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试商品",
    "price": 10000,
    "stock": 100,
    "seckill_stock": 50,
    "status": 2,
    "start_time": "2024-01-01T10:00:00Z",
    "end_time": "2024-12-31T23:59:59Z"
  }'

# 2. 检查Redis中的库存
redis-cli GET "seckill:stock:1"
# 应该返回: 50
```

### 2. RabbitMQ消息确认机制

```bash
# 1. 查看Worker日志
tail -f worker.log

# 2. 发送秒杀请求
# 3. 观察日志中是否有消息确认的记录
# 4. 停止Worker，发送消息
# 5. 重启Worker，消息应该被重新处理
```

### 3. Worker失败时的库存回滚

```bash
# 1. 停止MySQL服务（模拟数据库错误）
sudo systemctl stop mysql

# 2. 发送秒杀请求
# 3. Worker处理失败
# 4. 检查Redis库存是否回滚
redis-cli GET "seckill:stock:1"

# 5. 恢复MySQL服务
sudo systemctl start mysql
```

### 4. 前端秒杀结果展示

1. 打开浏览器访问商品详情页
2. 点击"立即秒杀"按钮
3. 观察是否有轮询查询结果
4. 观察是否显示秒杀成功/失败信息

### 5. 库存一致性检查

```bash
# 1. 运行库存同步服务
go run ./cmd/stock-sync

# 2. 手动修改Redis库存
redis-cli SET "seckill:stock:1" 999

# 3. 等待5分钟或手动触发检查
# 4. 检查Redis库存是否被修复为MySQL中的值
```

## 故障排查

### 问题1: 连接被拒绝

**错误**: `dial tcp 127.0.0.1:8080: connect: connection refused`

**解决**: 确保Web服务已启动

```bash
sudo systemctl status goseckill-web
# 或
ps aux | grep "cmd/web"
```

### 问题2: 认证失败

**错误**: `登录失败: invalid password`

**解决**: 确保测试用户存在，或先注册

### 问题3: 商品不存在

**错误**: `没有可用商品`

**解决**: 确保数据库中有商品数据

```bash
# 通过Admin接口创建商品
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试商品",
    "price": 10000,
    "stock": 100,
    "seckill_stock": 50,
    "status": 2,
    "start_time": "2024-01-01T10:00:00Z",
    "end_time": "2024-12-31T23:59:59Z"
  }'
```

### 问题4: 限流测试未生效

**原因**: 限流配置可能较宽松

**解决**: 调整限流参数或增加请求频率

## 测试覆盖

- ✅ API接口测试
- ✅ 功能逻辑测试
- ⚠️ 需要手动验证的功能（4项）
  - 库存自动同步
  - 消息确认机制
  - Worker失败回滚
  - 库存一致性检查

## 扩展测试

可以扩展测试程序添加：

1. **压力测试**: 使用goroutine并发测试
2. **集成测试**: 测试完整流程
3. **性能测试**: 测量响应时间
4. **边界测试**: 测试边界条件

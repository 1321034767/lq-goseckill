# seckill-worker 服务说明

## 功能概述

`seckill-worker` 是一个**后台异步订单处理服务**，负责从 RabbitMQ 消息队列中消费秒杀订单消息，并完成实际的订单创建和库存扣减操作。

## 核心作用

### 1. **异步处理秒杀订单**
- Web 服务接收到秒杀请求后，只负责：
  - 验证秒杀路径
  - 在 Redis 中预减库存（原子操作）
  - 将订单请求发送到 RabbitMQ 队列
  - 立即返回 "queued"（已排队）响应给用户

- Worker 服务负责：
  - 从队列中消费订单消息
  - 在 MySQL 中扣减商品库存
  - 创建订单记录
  - 设置秒杀成功标记（实现幂等性）

### 2. **削峰填谷，提升系统吞吐量**
- **高并发场景**：大量用户同时秒杀时，Web 服务可以快速响应，不需要等待数据库操作
- **异步处理**：订单创建操作在后台异步执行，不会阻塞用户请求
- **数据库保护**：避免高并发时直接操作 MySQL 导致数据库压力过大

### 3. **库存回滚机制**
- 如果订单创建失败（商品不存在、库存不足、数据库错误等）
- Worker 会自动回滚 Redis 中的库存（使用 `INCR` 操作）
- 确保 Redis 和 MySQL 的库存数据一致性

### 4. **消息确认机制**
- 使用**手动确认模式**（`auto-ack=false`）
- 订单创建成功后才 `Ack` 消息
- 订单创建失败时 `Nack` 消息并重新入队，保证消息不丢失

### 5. **幂等性保证**
- 订单创建成功后，在 Redis 中设置成功标记
- 标记格式：`seckill:succ:{userID}:{productID}`
- 有效期：24小时
- 用于防止用户重复秒杀

## 工作流程

```
用户发起秒杀请求
    ↓
Web 服务 (router.go)
    ├─ 验证秒杀路径
    ├─ Redis 预减库存 (DECR)
    └─ 发送消息到 RabbitMQ 队列
    ↓
立即返回 "queued" 给用户
    ↓
seckill-worker 消费消息
    ├─ 查询商品信息
    ├─ 检查库存
    ├─ MySQL 扣减库存
    ├─ 创建订单
    ├─ 设置成功标记
    └─ Ack 消息
```

## 是否必须运行？

### ✅ **必须运行**（如果使用秒杀功能）

**原因：**
1. **秒杀功能依赖**：Web 服务只负责将秒杀请求放入队列，**不会创建订单**
2. **订单无法创建**：如果 Worker 不运行，消息会堆积在队列中，订单永远不会被创建
3. **库存不一致**：Redis 中的库存已被预减，但 MySQL 中的库存没有扣减，导致数据不一致

### ⚠️ **可以不运行**（如果只使用普通购买功能）

**原因：**
- 普通购买接口 (`/api/purchase`) 是同步处理的，不依赖 Worker
- 普通购买直接在 Web 服务中完成订单创建，不需要消息队列

## 运行方式

### 1. 直接运行
```bash
go run ./cmd/seckill-worker
```

### 2. 编译后运行
```bash
go build -o seckill-worker ./cmd/seckill-worker
./seckill-worker
```

### 3. 后台运行（生产环境）
```bash
nohup ./seckill-worker > worker.log 2>&1 &
```

## 依赖服务

Worker 需要以下服务正常运行：

1. **MySQL** - 存储商品和订单数据
2. **Redis** - 回滚库存、设置成功标记
3. **RabbitMQ** - 消息队列服务

## 日志输出

Worker 会输出以下日志：
- `seckill worker started, waiting for messages...` - 启动成功
- `create order success, order_id=xxx` - 订单创建成功
- `rolled back redis stock for product xxx` - 库存回滚
- `failed to ...` - 各种错误信息

## 监控指标

Worker 会记录以下监控指标（通过 `service.GetMonitor()`）：
- `RecordWorkerProcessed()` - 成功处理的订单数
- `RecordWorkerFailed()` - 失败的订单数
- `RecordDBError()` - 数据库错误数

## 注意事项

1. **消息堆积**：如果 Worker 停止运行，消息会堆积在队列中，重启后会继续处理
2. **多实例部署**：可以运行多个 Worker 实例来提高处理能力（RabbitMQ 会自动分发消息）
3. **错误处理**：Worker 失败时会回滚库存并重新入队消息，确保数据一致性
4. **幂等性**：即使消息重复处理，由于有成功标记，不会创建重复订单

## 总结

**seckill-worker 是秒杀系统的核心组件之一**，负责实际的订单创建和库存扣减。如果使用秒杀功能，**必须运行 Worker 服务**，否则秒杀请求只会进入队列而不会创建订单。

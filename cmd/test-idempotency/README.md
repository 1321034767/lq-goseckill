# 幂等性测试程序

## 功能说明

此测试程序用于验证秒杀系统的幂等性保证功能，确保同一用户对同一商品只能成功秒杀一次。

## 测试流程

1. **注册/登录用户** - 获取JWT Token
2. **获取秒杀路径** - 获取动态生成的秒杀地址
3. **第一次秒杀** - 应该成功，返回 "queued"
4. **等待处理** - 等待Worker处理订单（3秒）
5. **再次获取路径** - 模拟用户刷新页面获取新路径
6. **第二次秒杀** - 应该被拒绝，返回 "duplicate seckill"

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

### 2. 确保商品存在且有库存

确保商品ID 1存在且有秒杀库存。可以通过Admin后台管理界面或直接操作数据库。

### 3. 初始化Redis库存

如果Redis中没有库存数据，需要先初始化：

```bash
# 可以通过Admin接口更新商品，或者直接操作Redis
redis-cli SET seckill:stock:1 10
```

## 运行测试

```bash
cd cmd/test-idempotency
go run main.go
```

## 预期结果

```
=== 幂等性测试程序 ===

步骤1: 注册/登录用户...
✅ 登录成功，Token: eyJhbGciOiJIUzI1NiIs...

步骤2: 获取秒杀路径...
✅ 获取路径成功: abc123def456...

步骤3: 第一次秒杀请求...
✅ 第一次秒杀响应: code=0, msg=queued

等待3秒，确保Worker处理完订单...

步骤4: 再次获取秒杀路径...
✅ 获取新路径: xyz789uvw012...

步骤5: 第二次秒杀请求（应该被拒绝）...
响应: code=400, msg=duplicate seckill

=== 测试结果 ===
✅ 幂等性测试通过！
   - 第一次秒杀成功
   - 第二次秒杀被正确拒绝（duplicate seckill）
```

## 验证Redis中的成功标记

测试完成后，可以验证Redis中是否设置了成功标记：

```bash
redis-cli GET "seckill:succ:1:1"
# 应该返回: "1"

redis-cli TTL "seckill:succ:1:1"
# 应该返回剩余过期时间（秒），约86400秒（24小时）
```

## 故障排查

### 问题1: 连接被拒绝

**错误**: `dial tcp 127.0.0.1:8080: connect: connection refused`

**解决**: 确保Web服务已启动，监听8080端口

```bash
sudo systemctl status goseckill-web
# 或
ps aux | grep "cmd/web"
```

### 问题2: 第二次秒杀没有被拒绝

**可能原因**:
1. Worker没有成功处理订单（检查Worker日志）
2. Redis连接失败（检查Redis服务状态）
3. Worker中没有设置成功标记（检查代码）

**排查步骤**:
```bash
# 检查Worker日志
tail -f /var/log/goseckill-worker.log
# 或
journalctl -u goseckill-worker -f

# 检查Redis连接
redis-cli PING

# 手动检查成功标记
redis-cli KEYS "seckill:succ:*"
```

### 问题3: 商品库存不足

**错误**: `stock empty`

**解决**: 确保商品有足够的秒杀库存

```bash
# 检查MySQL中的库存
mysql -u goseckill -p goseckill123 -e "SELECT id, seckill_stock FROM products WHERE id=1;"

# 初始化Redis库存
redis-cli SET seckill:stock:1 10
```

## 技术实现说明

### 幂等性保证机制

1. **秒杀前检查**: 在 `Seckill()` 方法中，使用 `EXISTS` 检查Redis中是否存在成功标记
   - Key格式: `seckill:succ:{userID}:{productID}`
   - 如果存在，直接返回 "duplicate seckill"

2. **设置成功标记**: 在Worker成功创建订单后，设置Redis成功标记
   - 使用 `SETEX` 设置，有效期24小时
   - 即使Redis设置失败，订单已创建，不影响订单创建流程

3. **标记有效期**: 24小时（86400秒）
   - 防止用户重复秒杀
   - 过期后可以再次秒杀（如果需要）

### 代码位置

- **检查逻辑**: `internal/service/seckill_service.go:81-87`
- **设置标记**: `cmd/seckill-worker/main.go:80-88`

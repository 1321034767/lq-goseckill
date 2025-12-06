# 手动测试指南

某些功能无法通过自动化测试完全验证，需要手动测试。以下是详细的手动测试步骤。

## 1. 库存自动同步机制测试

### 测试步骤

1. **查看当前Redis库存**
   ```bash
   redis-cli GET "seckill:stock:1"
   ```

2. **通过Admin接口更新商品库存**
   ```bash
   curl -X PUT http://localhost:8081/api/products/1 \
     -H "Content-Type: application/json" \
     -d '{
       "name": "测试商品",
       "price": 10000,
       "stock": 100,
       "seckill_stock": 80,
       "status": 2,
       "start_time": "2024-01-01T10:00:00Z",
       "end_time": "2024-12-31T23:59:59Z"
     }'
   ```

3. **再次检查Redis库存**
   ```bash
   redis-cli GET "seckill:stock:1"
   # 应该返回: 80
   ```

### 预期结果

- Redis中的库存应该自动更新为80
- 如果返回80，说明自动同步机制工作正常

---

## 2. RabbitMQ消息确认机制测试

### 测试步骤

1. **启动Worker并查看日志**
   ```bash
   go run ./cmd/seckill-worker 2>&1 | tee worker.log
   ```

2. **发送秒杀请求**
   ```bash
   # 先登录获取token
   TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass"}' \
     | jq -r '.data.token')
   
   # 获取秒杀路径
   PATH=$(curl -s -X GET http://localhost:8080/api/seckill/1/path \
     -H "Authorization: $TOKEN" \
     | jq -r '.data.path')
   
   # 发起秒杀
   curl -X POST http://localhost:8080/api/seckill/1/$PATH \
     -H "Authorization: $TOKEN"
   ```

3. **观察Worker日志**
   - 应该看到 "create order success" 日志
   - 应该看到 "set seckill success mark" 日志

4. **测试消息重试**
   - 停止Worker（Ctrl+C）
   - 发送秒杀请求
   - 重启Worker
   - 消息应该被重新处理

### 预期结果

- Worker处理成功后，消息被确认（Ack）
- Worker处理失败时，消息被拒绝并重新入队（Nack）
- 重启Worker后，未确认的消息会被重新处理

---

## 3. Worker失败时的库存回滚测试

### 测试步骤

1. **记录当前Redis库存**
   ```bash
   redis-cli GET "seckill:stock:1"
   # 假设返回: 50
   ```

2. **模拟数据库错误**
   ```bash
   # 临时停止MySQL（谨慎操作）
   sudo systemctl stop mysql
   ```

3. **发送秒杀请求**
   ```bash
   # 使用上面的TOKEN和PATH
   curl -X POST http://localhost:8080/api/seckill/1/$PATH \
     -H "Authorization: $TOKEN"
   ```

4. **检查Redis库存**
   ```bash
   redis-cli GET "seckill:stock:1"
   # 应该仍然是50（库存已回滚）
   ```

5. **恢复MySQL**
   ```bash
   sudo systemctl start mysql
   ```

### 预期结果

- Worker处理失败时，Redis库存应该回滚
- 库存数量应该恢复到秒杀前的值

---

## 4. 前端秒杀结果展示测试

### 测试步骤

1. **打开浏览器**
   - 访问: http://localhost:8080/product/1

2. **登录**
   - 点击登录，使用测试账号登录

3. **发起秒杀**
   - 点击"立即秒杀"按钮
   - 观察页面变化

4. **观察结果**
   - 应该看到"秒杀请求已提交，正在处理中..."
   - 几秒后应该看到"秒杀成功！订单号：XXX"
   - 或者看到"查询超时，请稍后查看订单列表"

### 预期结果

- 秒杀按钮变为"秒杀中..."状态
- 自动轮询查询结果
- 显示秒杀成功/失败信息
- 库存实时更新

---

## 5. 库存一致性检查测试

### 测试步骤

1. **启动库存同步服务**
   ```bash
   go run ./cmd/stock-sync
   ```

2. **手动修改Redis库存（制造不一致）**
   ```bash
   redis-cli SET "seckill:stock:1" 999
   ```

3. **等待同步服务检查（或手动触发）**
   - 同步服务每5分钟检查一次
   - 或者修改代码中的间隔时间

4. **检查日志**
   - 应该看到 "商品 1: 库存不一致" 的日志
   - 应该看到 "已修复库存不一致" 的日志

5. **验证Redis库存**
   ```bash
   redis-cli GET "seckill:stock:1"
   # 应该返回MySQL中的值，而不是999
   ```

### 预期结果

- 检测到库存不一致
- 自动修复不一致（以MySQL为准）
- 记录修复日志

---

## 6. 限流功能测试

### 测试步骤

1. **快速发送多个请求**
   ```bash
   # 获取token和path
   TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"testpass"}' \
     | jq -r '.data.token')
   
   PATH=$(curl -s -X GET http://localhost:8080/api/seckill/1/path \
     -H "Authorization: $TOKEN" \
     | jq -r '.data.path')
   
   # 快速发送20个请求
   for i in {1..20}; do
     curl -s -X POST http://localhost:8080/api/seckill/1/$PATH \
       -H "Authorization: $TOKEN" \
       | jq -r '.code, .msg'
     echo "---"
   done
   ```

2. **观察响应**
   - 前几个请求应该成功（code=0）
   - 后续请求应该被限流（code=429）

### 预期结果

- 部分请求成功
- 超过限流的请求返回429错误
- 错误信息: "请求过于频繁，请稍后再试"

---

## 7. 监控功能测试

### 测试步骤

1. **发送一些秒杀请求**

2. **查看监控统计**
   ```bash
   curl http://localhost:8081/api/monitor/stats | jq
   ```

3. **观察统计数据**
   - 错误统计（Redis、MQ、DB等）
   - 性能统计（请求数、成功率等）
   - 最后事件时间

### 预期结果

- 返回完整的监控数据
- 统计数据准确反映系统状态

---

## 测试检查清单

- [ ] 库存自动同步机制
- [ ] RabbitMQ消息确认机制
- [ ] Worker失败时的库存回滚
- [ ] 前端秒杀结果展示
- [ ] 库存一致性检查
- [ ] 限流功能
- [ ] 监控功能

## 注意事项

1. **数据库操作**: 停止MySQL会影响其他服务，建议在测试环境操作
2. **Redis操作**: 修改Redis库存后记得恢复
3. **Worker日志**: 保持Worker运行以观察日志
4. **测试数据**: 使用测试商品，避免影响生产数据

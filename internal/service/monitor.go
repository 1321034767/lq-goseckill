package service

import (
	"sync"
	"time"
)

// Monitor 监控服务，用于统计错误和性能指标
type Monitor struct {
	mu sync.RWMutex

	// 错误统计
	RedisErrors      int64
	MQErrors         int64
	DBErrors         int64
	SeckillErrors    int64
	WorkerErrors     int64

	// 性能统计
	SeckillRequests  int64
	SeckillSuccess   int64
	WorkerProcessed  int64
	WorkerFailed     int64

	// 时间统计
	LastRedisError   time.Time
	LastMQError      time.Time
	LastDBError      time.Time
	LastSeckillTime  time.Time
	LastWorkerTime   time.Time
}

var globalMonitor = &Monitor{}

// GetMonitor 获取全局监控实例
func GetMonitor() *Monitor {
	return globalMonitor
}

// RecordRedisError 记录Redis错误
func (m *Monitor) RecordRedisError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RedisErrors++
	m.LastRedisError = time.Now()
}

// RecordMQError 记录MQ错误
func (m *Monitor) RecordMQError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MQErrors++
	m.LastMQError = time.Now()
}

// RecordDBError 记录数据库错误
func (m *Monitor) RecordDBError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DBErrors++
	m.LastDBError = time.Now()
}

// RecordSeckillRequest 记录秒杀请求
func (m *Monitor) RecordSeckillRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SeckillRequests++
	m.LastSeckillTime = time.Now()
}

// RecordSeckillSuccess 记录秒杀成功
func (m *Monitor) RecordSeckillSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SeckillSuccess++
}

// RecordSeckillError 记录秒杀错误
func (m *Monitor) RecordSeckillError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SeckillErrors++
}

// RecordWorkerProcessed 记录Worker处理成功
func (m *Monitor) RecordWorkerProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WorkerProcessed++
	m.LastWorkerTime = time.Now()
}

// RecordWorkerFailed 记录Worker处理失败
func (m *Monitor) RecordWorkerFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WorkerFailed++
	m.WorkerErrors++
}

// GetStats 获取统计信息
func (m *Monitor) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	successRate := float64(0)
	if m.SeckillRequests > 0 {
		successRate = float64(m.SeckillSuccess) / float64(m.SeckillRequests) * 100
	}

	workerSuccessRate := float64(0)
	totalWorker := m.WorkerProcessed + m.WorkerFailed
	if totalWorker > 0 {
		workerSuccessRate = float64(m.WorkerProcessed) / float64(totalWorker) * 100
	}

	return map[string]interface{}{
		"errors": map[string]interface{}{
			"redis":   m.RedisErrors,
			"mq":      m.MQErrors,
			"db":      m.DBErrors,
			"seckill": m.SeckillErrors,
			"worker":  m.WorkerErrors,
		},
		"performance": map[string]interface{}{
			"seckill_requests":    m.SeckillRequests,
			"seckill_success":     m.SeckillSuccess,
			"seckill_success_rate": successRate,
			"worker_processed":    m.WorkerProcessed,
			"worker_failed":       m.WorkerFailed,
			"worker_success_rate": workerSuccessRate,
		},
		"last_events": map[string]interface{}{
			"redis_error":  m.LastRedisError,
			"mq_error":     m.LastMQError,
			"db_error":     m.LastDBError,
			"last_seckill":  m.LastSeckillTime,
			"last_worker":   m.LastWorkerTime,
		},
	}
}

// Reset 重置统计（用于测试或定期清理）
func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RedisErrors = 0
	m.MQErrors = 0
	m.DBErrors = 0
	m.SeckillErrors = 0
	m.WorkerErrors = 0
	m.SeckillRequests = 0
	m.SeckillSuccess = 0
	m.WorkerProcessed = 0
	m.WorkerFailed = 0
}

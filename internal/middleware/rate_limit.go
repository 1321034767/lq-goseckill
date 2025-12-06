package middleware

import (
	"sync"
	"time"

	"github.com/kataras/iris/v12"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	capacity     int64         // 桶容量
	tokens       int64         // 当前令牌数
	refillRate   int64         // 每秒补充的令牌数
	lastRefill   time.Time     // 上次补充时间
	mu           sync.Mutex    // 互斥锁
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate
	if tokensToAdd > 0 {
		tb.tokens = tb.tokens + tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// 检查是否有可用令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(bucket *TokenBucket) iris.Handler {
	return func(ctx iris.Context) {
		if !bucket.Allow() {
			ctx.StopWithJSON(429, iris.Map{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}
		ctx.Next()
	}
}

// 全局限流器（可根据需要创建多个）
var (
	seckillRateLimiter = NewTokenBucket(10, 5) // 容量10，每秒补充5个令牌（更严格的限流用于测试）
)

// SeckillRateLimit 秒杀接口限流
func SeckillRateLimit() iris.Handler {
	return RateLimitMiddleware(seckillRateLimiter)
}

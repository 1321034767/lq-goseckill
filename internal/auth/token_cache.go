package auth

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	radix "github.com/mediocregopher/radix/v3"
)

// TokenCache 基于一致性哈希的 JWT 解析结果缓存，加速分布式鉴权
type TokenCache struct {
	redis radix.Client
	ring  *ConsistentHashRing
	ttl   time.Duration
}

// NewTokenCache 构建缓存器
func NewTokenCache(redis radix.Client, ring *ConsistentHashRing, ttl time.Duration) *TokenCache {
	if ring == nil {
		ring = NewConsistentHashRing(nil, 0)
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &TokenCache{
		redis: redis,
		ring:  ring,
		ttl:   ttl,
	}
}

func (c *TokenCache) cacheKey(token string) string {
	node := c.ring.GetNode(token)
	sum := sha1.Sum([]byte(token))
	return fmt.Sprintf("auth:jwt:%s:%s", node, hex.EncodeToString(sum[:]))
}

// Get 尝试命中缓存的 claims
func (c *TokenCache) Get(ctx context.Context, token string) (*Claims, bool, error) {
	if c.redis == nil {
		return nil, false, nil
	}
	key := c.cacheKey(token)
	var raw string
	if err := c.redis.Do(radix.Cmd(&raw, "GET", key)); err != nil {
		return nil, false, err
	}
	if raw == "" {
		return nil, false, nil
	}
	var claims Claims
	if err := json.Unmarshal([]byte(raw), &claims); err != nil {
		// 数据损坏，清理后走正常解析
		_ = c.redis.Do(radix.Cmd(nil, "DEL", key))
		return nil, false, nil
	}
	return &claims, true, nil
}

// Set 缓存解析结果
func (c *TokenCache) Set(ctx context.Context, token string, claims *Claims) error {
	if c.redis == nil || claims == nil {
		return nil
	}
	key := c.cacheKey(token)
	body, _ := json.Marshal(claims)
	if err := c.redis.Do(radix.FlatCmd(nil, "SETEX", key, int64(c.ttl/time.Second), body)); err != nil {
		return err
	}
	return nil
}

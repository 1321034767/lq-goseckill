package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/goseckill/internal/auth"
	"github.com/example/goseckill/internal/config"
	redisInfra "github.com/example/goseckill/internal/infra/redis"
)

// 一个简单的自测程序：演示一致性哈希节点选择 + JWT 解析与缓存命中
func main() {
	cfg := config.DefaultConfig()
	rdb := redisInfra.Init(&cfg.Redis)

	// 构建一致性哈希环与缓存
	ring := auth.NewConsistentHashRing(cfg.Auth.Nodes, cfg.Auth.HashReplicas)
	cache := auth.NewTokenCache(rdb, ring, time.Duration(cfg.Auth.TokenCacheTTLSeconds)*time.Second)

	token, err := auth.GenerateToken(&cfg.JWT, 12345, "demo-user")
	if err != nil {
		log.Fatalf("generate token failed: %v", err)
	}

	ctx := context.Background()
	node := ring.GetNode(token)
	fmt.Println("选择的鉴权节点:", node)
	fmt.Println("生成的 JWT:", token)

	// 第一次解析（未命中缓存）
	claims, err := auth.ParseToken(&cfg.JWT, token)
	if err != nil {
		log.Fatalf("parse token failed: %v", err)
	}
	_ = cache.Set(ctx, token, claims)
	fmt.Printf("首次解析写入缓存，user_id=%d, username=%s\n", claims.UserID, claims.Username)

	// 第二次解析（命中缓存）
	cached, ok, err := cache.Get(ctx, token)
	if err != nil {
		log.Fatalf("cache get failed: %v", err)
	}
	fmt.Printf("二次命中缓存=%v, user_id=%d, username=%s\n", ok, cached.UserID, cached.Username)
}

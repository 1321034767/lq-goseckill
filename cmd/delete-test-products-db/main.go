package main

import (
	"context"
	"fmt"
	"log"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/repository/mysql"
)

func main() {
	cfg := config.DefaultConfig()
	db := mysql.Init(&cfg.MySQL)
	productRepo := mysql.NewProductRepository(db)

	ctx := context.Background()

	// 获取所有商品
	products, err := productRepo.ListAll(ctx)
	if err != nil {
		log.Fatalf("获取商品列表失败: %v", err)
	}

	fmt.Printf("📋 找到 %d 个商品\n\n", len(products))

	// 查找测试商品（名称包含"测试"的商品）
	testProducts := []*product.Product{}
	for _, p := range products {
		if p.Name == "测试秒杀商品" {
			testProducts = append(testProducts, p)
		}
	}

	if len(testProducts) == 0 {
		fmt.Println("ℹ️  没有找到测试商品")
		return
	}

	fmt.Printf("🔍 找到 %d 个测试商品:\n", len(testProducts))
	for _, p := range testProducts {
		fmt.Printf("  - ID: %d, 名称: %s\n", p.ID, p.Name)
	}

	// 删除前两个测试商品
	deleteCount := 2
	if len(testProducts) < deleteCount {
		deleteCount = len(testProducts)
	}

	fmt.Printf("\n🗑️  准备删除前 %d 个测试商品...\n\n", deleteCount)

	successCount := 0
	failCount := 0

	for i := 0; i < deleteCount; i++ {
		p := testProducts[i]
		if err := productRepo.Delete(ctx, p.ID); err != nil {
			fmt.Printf("❌ 商品 %d (%s) 删除失败: %v\n", p.ID, p.Name, err)
			failCount++
			continue
		}
		fmt.Printf("✅ 商品 %d (%s) 删除成功\n", p.ID, p.Name)
		successCount++
	}

	fmt.Printf("\n📊 总结: 成功删除 %d 个, 失败 %d 个\n", successCount, failCount)
}

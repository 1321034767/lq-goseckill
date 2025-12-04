package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Product struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Status       int    `json:"status"`
}

type ApiResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data []Product `json:"data"`
}

func main() {
	adminURL := "http://localhost:8081/api/products"
	client := &http.Client{}

	fmt.Println("🧪 测试商品ID范围...")
	fmt.Println("=" + strings.Repeat("=", 60))

	// 获取所有商品
	resp, err := client.Get(adminURL)
	if err != nil {
		fmt.Printf("❌ 获取商品列表失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ HTTP状态码: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		fmt.Printf("响应内容: %s\n", string(body))
		os.Exit(1)
	}

	if apiResp.Code != 0 {
		fmt.Printf("❌ API返回错误: code=%d, msg=%s\n", apiResp.Code, apiResp.Msg)
		os.Exit(1)
	}

	products := apiResp.Data
	fmt.Printf("\n📋 找到 %d 个商品\n\n", len(products))

	if len(products) == 0 {
		fmt.Println("⚠️  没有商品")
		os.Exit(1)
	}

	// 分析商品ID
	minID := products[0].ID
	maxID := products[0].ID
	allIDs := []int64{}

	for _, p := range products {
		allIDs = append(allIDs, p.ID)
		if p.ID < minID {
			minID = p.ID
		}
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	fmt.Println("📊 商品ID统计:")
	fmt.Printf("   最小ID: %d\n", minID)
	fmt.Printf("   最大ID: %d\n", maxID)
	fmt.Printf("   ID范围: %d - %d\n", minID, maxID)
	fmt.Printf("   商品数量: %d\n", len(products))

	// 检查ID是否连续
	fmt.Println("\n🔍 检查ID连续性:")
	expectedIDs := make(map[int64]bool)
	for i := int64(1); i <= 12; i++ {
		expectedIDs[i] = false
	}

	for _, id := range allIDs {
		if id >= 1 && id <= 12 {
			expectedIDs[id] = true
		}
	}

	missingIDs := []int64{}
	for id, found := range expectedIDs {
		if !found {
			missingIDs = append(missingIDs, id)
		}
	}

	extraIDs := []int64{}
	for _, id := range allIDs {
		if id < 1 || id > 12 {
			extraIDs = append(extraIDs, id)
		}
	}

	// 验证结果
	fmt.Println("\n✅ 验证结果:")
	success := true

	if minID == 1 && maxID == 12 && len(products) == 12 {
		fmt.Println("   ✅ 商品ID范围正确: 1-12")
		fmt.Println("   ✅ 商品数量正确: 12个")
	} else {
		success = false
		fmt.Printf("   ❌ 商品ID范围不正确: %d-%d (期望: 1-12)\n", minID, maxID)
		if len(products) != 12 {
			fmt.Printf("   ❌ 商品数量不正确: %d (期望: 12)\n", len(products))
		}
	}

	if len(missingIDs) > 0 {
		success = false
		fmt.Printf("   ❌ 缺少ID: %v\n", missingIDs)
	}

	if len(extraIDs) > 0 {
		success = false
		fmt.Printf("   ❌ 超出范围的ID: %v\n", extraIDs)
	}

	// 显示所有商品ID
	fmt.Println("\n📝 所有商品ID列表:")
	for _, p := range products {
		status := "✅"
		if p.ID < 1 || p.ID > 12 {
			status = "❌"
		}
		fmt.Printf("   %s ID: %2d - %s\n", status, p.ID, p.Name)
	}

	// 图片映射验证
	fmt.Println("\n🖼️  图片映射验证:")
	for _, p := range products {
		var expectedImg int64
		if p.ID >= 1 && p.ID <= 12 {
			expectedImg = p.ID
		} else {
			expectedImg = ((p.ID - 1) % 12) + 1
		}
		fmt.Printf("   商品ID %2d → 图片 product_%d.jpg\n", p.ID, expectedImg)
	}

	// 最终结果
	fmt.Println("\n" + strings.Repeat("=", 60))
	if success {
		fmt.Println("🎉 测试通过！商品ID确实是 1-12")
		os.Exit(0)
	} else {
		fmt.Println("❌ 测试失败！商品ID不是 1-12")
		fmt.Println("\n💡 建议运行以下命令重置:")
		fmt.Println("   go run ./cmd/reset-products-id/main.go")
		os.Exit(1)
	}
}

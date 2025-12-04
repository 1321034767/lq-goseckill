package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ProductRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Category     string `json:"category"` // 分类：men(男士)、women(女士)、accessories(饰品)
	Status       int    `json:"status"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

type Product struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Category     string `json:"category"`
	Status       int    `json:"status"`
}

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func main() {
	adminURL := "http://localhost:8081/api"
	client := &http.Client{}

	fmt.Println("🔄 更新现有商品的分类...")
	fmt.Println("=" + strings.Repeat("=", 60))

	// 步骤1: 获取所有商品列表
	fmt.Println("\n[1/2] 获取所有商品列表...")
	resp, err := client.Get(adminURL + "/products")
	if err != nil {
		fmt.Printf("❌ 获取商品列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var listResp struct {
		Code int       `json:"code"`
		Data []Product `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		fmt.Printf("❌ 解析商品列表失败: %v\n", err)
		return
	}

	if listResp.Code != 0 {
		fmt.Printf("❌ 获取商品列表失败: code=%d\n", listResp.Code)
		return
	}

	fmt.Printf("📋 找到 %d 个商品\n", len(listResp.Data))

	if len(listResp.Data) == 0 {
		fmt.Println("ℹ️  没有商品需要更新")
		return
	}

	// 商品分类映射（根据商品名称和ID）
	type CategoryMapping struct {
		ID       int64
		Name     string
		Category string
	}

	categoryMap := map[int64]string{
		1:  "men",        // Joeby Tailored Trouser
		2:  "men",        // Denim Hooded
		3:  "women",      // Mint Maxi Dress
		4:  "women",      // White Flounce Dress
		5:  "men",        // Classic White Shirt
		6:  "men",        // Casual Denim Jacket
		7:  "women",      // Elegant Blouse
		8:  "men",        // Stylish T-Shirt
		9:  "accessories", // Men's Belt
		10: "accessories", // Sport Hi Adidas
		11: "women",      // Leather Handbag (改为女士)
		12: "accessories", // Designer Sunglasses
	}

	// 步骤2: 更新每个商品的分类
	fmt.Println("\n[2/2] 更新商品分类...")
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	successCount := 0
	failCount := 0

	for _, p := range listResp.Data {
		category, exists := categoryMap[p.ID]
		if !exists {
			// 如果ID不在映射表中，根据商品名称判断
			if contains(p.Name, []string{"Belt", "Adidas", "Handbag", "Sunglasses"}) {
				category = "accessories"
			} else if contains(p.Name, []string{"Dress", "Blouse"}) {
				category = "women"
			} else {
				category = "men"
			}
		}

		// 如果分类已经是正确的，跳过
		if p.Category == category {
			fmt.Printf("ℹ️  商品 %d (%s) 分类已是 %s，跳过\n", p.ID, p.Name, category)
			continue
		}

		req := ProductRequest{
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			Stock:        p.Stock,
			SeckillStock: p.SeckillStock,
			Category:     category,
			Status:       p.Status,
			StartTime:    startTime,
			EndTime:      endTime,
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("❌ 商品 %d: JSON 序列化失败: %v\n", p.ID, err)
			failCount++
			continue
		}

		updateURL := fmt.Sprintf("%s/products/%d", adminURL, p.ID)
		httpReq, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ 商品 %d: 创建请求失败: %v\n", p.ID, err)
			failCount++
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")

		updateResp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 请求失败: %v\n", p.ID, err)
			failCount++
			continue
		}
		defer updateResp.Body.Close()

		updateBody, err := io.ReadAll(updateResp.Body)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 读取响应失败: %v\n", p.ID, err)
			failCount++
			continue
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(updateBody, &apiResp); err != nil {
			fmt.Printf("❌ 商品 %d: JSON 解析失败: %v\n", p.ID, err)
			failCount++
			continue
		}

		if apiResp.Code == 0 {
			fmt.Printf("✅ 商品 %d (%s) 分类更新为 %s\n", p.ID, p.Name, category)
			successCount++
		} else {
			fmt.Printf("❌ 商品 %d (%s) 更新失败: %s\n", p.ID, p.Name, apiResp.Msg)
			failCount++
		}
	}

	fmt.Printf("\n📊 更新总结: 成功 %d 个, 失败 %d 个\n", successCount, failCount)

	if successCount > 0 {
		fmt.Println("\n✅ 商品分类更新完成！")
		fmt.Println("   现在可以点击分类按钮查看对应分类的商品了")
	}
}

func contains(s string, list []string) bool {
	sLower := strings.ToLower(s)
	for _, item := range list {
		if strings.Contains(sLower, strings.ToLower(item)) {
			return true
		}
	}
	return false
}

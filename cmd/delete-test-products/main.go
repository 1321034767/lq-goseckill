package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func main() {
	adminURL := "http://localhost:8081/api"
	client := &http.Client{}

	// 获取所有商品列表
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
		Code int        `json:"code"`
		Data []Product  `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		fmt.Printf("❌ 解析商品列表失败: %v\n", err)
		return
	}

	if listResp.Code != 0 {
		fmt.Printf("❌ 获取商品列表失败: code=%d\n", listResp.Code)
		return
	}

	fmt.Printf("📋 找到 %d 个商品\n\n", len(listResp.Data))

	// 查找测试商品（名称包含"测试"的商品）
	testProducts := []Product{}
	for _, p := range listResp.Data {
		if strings.Contains(p.Name, "测试") || strings.Contains(strings.ToLower(p.Name), "test") {
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
		deleteURL := fmt.Sprintf("%s/products/%d", adminURL, p.ID)
		
		req, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			fmt.Printf("❌ 商品 %d (%s): 创建请求失败: %v\n", p.ID, p.Name, err)
			failCount++
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 商品 %d (%s): 请求失败: %v\n", p.ID, p.Name, err)
			failCount++
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 商品 %d (%s): 读取响应失败: %v\n", p.ID, p.Name, err)
			failCount++
			continue
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			fmt.Printf("❌ 商品 %d (%s): JSON 解析失败: %v\n", p.ID, p.Name, err)
			fmt.Printf("   响应内容: %s\n", string(respBody))
			failCount++
			continue
		}

		if apiResp.Code == 0 {
			fmt.Printf("✅ 商品 %d (%s) 删除成功\n", p.ID, p.Name)
			successCount++
		} else {
			fmt.Printf("❌ 商品 %d (%s) 删除失败: %s\n", p.ID, p.Name, apiResp.Msg)
			failCount++
		}
	}

	fmt.Printf("\n📊 总结: 成功删除 %d 个, 失败 %d 个\n", successCount, failCount)
}

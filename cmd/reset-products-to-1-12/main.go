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

	fmt.Println("🔄 重置商品ID为1-12...")
	fmt.Println("=" + strings.Repeat("=", 50))

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
		fmt.Println("ℹ️  没有商品需要删除")
		fmt.Println("\n✅ 请运行以下命令添加商品（ID将自动从1开始）:")
		fmt.Println("   go run ./cmd/add-products/main.go")
		return
	}

	// 步骤2: 删除所有商品
	fmt.Println("\n[2/2] 删除所有商品...")
	successCount := 0
	failCount := 0

	for _, p := range listResp.Data {
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

	fmt.Printf("\n📊 删除总结: 成功 %d 个, 失败 %d 个\n", successCount, failCount)

	if successCount > 0 {
		fmt.Println("\n✅ 所有商品已删除！")
		fmt.Println("\n📝 下一步：运行以下命令添加12个商品（ID将自动从1开始）:")
		fmt.Println("   go run ./cmd/add-products/main.go")
		fmt.Println("\n这样新添加的商品ID将是 1, 2, 3, ..., 12")
	}
}

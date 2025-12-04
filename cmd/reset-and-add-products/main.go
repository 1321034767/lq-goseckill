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

type Product struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Status       int    `json:"status"`
}

type ProductRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Status       int    `json:"status"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
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
	fmt.Println("=" + strings.Repeat("=", 60))

	// 步骤1: 获取所有商品列表
	fmt.Println("\n[1/3] 获取所有商品列表...")
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

	// 步骤2: 删除所有商品
	if len(listResp.Data) > 0 {
		fmt.Println("\n[2/3] 删除所有现有商品...")
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
	} else {
		fmt.Println("\n[2/3] 没有商品需要删除")
	}

	// 步骤3: 添加12个新商品
	fmt.Println("\n[3/3] 添加12个新商品（ID将从1开始）...")

	type ProductInfo struct {
		Name        string
		Description string
		Price       int64
	}

	products := []ProductInfo{
		{"Joeby Tailored Trouser", "经典剪裁长裤，舒适面料，适合日常穿着。", 1799},  // product_1.jpg
		{"Denim Hooded", "牛仔连帽衫，时尚百搭，适合休闲场合。", 3000},              // product_2.jpg
		{"Mint Maxi Dress", "薄荷绿长款连衣裙，优雅大方，适合多种场合。", 1799},      // product_3.jpg
		{"White Flounce Dress", "白色荷叶边连衣裙，清新甜美，展现女性魅力。", 1599}, // product_4.jpg
		{"Classic White Shirt", "经典白色衬衫，百搭单品，职场必备。", 1999},        // product_5.jpg
		{"Casual Denim Jacket", "休闲牛仔夹克，经典款式，四季可穿。", 2500},        // product_6.jpg
		{"Elegant Blouse", "优雅女士衬衫，精致剪裁，展现知性美。", 2200},          // product_7.jpg
		{"Stylish T-Shirt", "时尚T恤，舒适面料，休闲百搭。", 1299},                // product_8.jpg
		{"Men's Belt", "男士皮带，真皮材质，经典设计。", 990},                      // product_9.jpg
		{"Sport Hi Adidas", "阿迪达斯运动鞋，舒适透气，适合运动健身。", 2900},      // product_10.jpg
		{"Leather Handbag", "真皮手提包，精致工艺，时尚优雅。", 3500},              // product_11.jpg
		{"Designer Sunglasses", "设计师太阳镜，防紫外线，时尚潮流。", 1800},        // product_12.jpg
	}

	// 设置秒杀时间：从现在开始，持续24小时
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	addSuccessCount := 0
	addFailCount := 0

	for i := 1; i <= 12; i++ {
		product := products[i-1]

		// 库存和秒杀库存根据价格合理设置
		stock := 50 + int64(i*5)        // 库存：55, 60, 65...
		seckillStock := 5 + int64(i/2)  // 秒杀库存：5, 5, 6, 6, 7, 7...

		req := ProductRequest{
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Stock:        stock,
			SeckillStock: seckillStock,
			Status:       1, // 1:正常
			StartTime:    startTime,
			EndTime:      endTime,
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("❌ 商品 %d: JSON 序列化失败: %v\n", i, err)
			addFailCount++
			continue
		}

		httpReq, err := http.NewRequest("POST", adminURL+"/products", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ 商品 %d: 创建请求失败: %v\n", i, err)
			addFailCount++
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 请求失败: %v\n", i, err)
			addFailCount++
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 读取响应失败: %v\n", i, err)
			addFailCount++
			continue
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			fmt.Printf("❌ 商品 %d: JSON 解析失败: %v\n", i, err)
			fmt.Printf("   响应内容: %s\n", string(respBody))
			addFailCount++
			continue
		}

		if apiResp.Code == 0 {
			// 尝试从响应中获取新创建的商品ID
			var productData map[string]interface{}
			if data, ok := apiResp.Data.(map[string]interface{}); ok {
				productData = data
			} else if dataBytes, err := json.Marshal(apiResp.Data); err == nil {
				json.Unmarshal(dataBytes, &productData)
			}

			productID := "?"
			if id, ok := productData["id"].(float64); ok {
				productID = fmt.Sprintf("%.0f", id)
			}

			fmt.Printf("✅ 商品 %d (%s) - ¥%.2f 添加成功 (ID: %s)\n", i, product.Name, float64(product.Price)/100, productID)
			addSuccessCount++
		} else {
			fmt.Printf("❌ 商品 %d (%s) 添加失败: %s\n", i, product.Name, apiResp.Msg)
			addFailCount++
		}
	}

	fmt.Printf("\n📊 添加总结: 成功 %d 个, 失败 %d 个\n", addSuccessCount, addFailCount)

	if addSuccessCount == 12 {
		fmt.Println("\n🎉 完成！商品ID现在是 1-12")
		fmt.Println("   商品ID 1 → 图片 product_1.jpg")
		fmt.Println("   商品ID 2 → 图片 product_2.jpg")
		fmt.Println("   ...")
		fmt.Println("   商品ID 12 → 图片 product_12.jpg")
	}
}

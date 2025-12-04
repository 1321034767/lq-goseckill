package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func main() {
	adminURL := "http://localhost:8081/api"
	
	// 商品信息列表（从 copy 文件夹中提取的商品名称和价格）
	// 价格单位：分（例如 $17.99 = 1799分）
	
	// 商品价格对照表（根据 copy/GoSecKill-main/web/server/views/static/index.html）
	// 价格单位：分（例如 $17.99 = 1799分 = ¥17.99）
	// 分类：men(男士)、women(女士)、accessories(饰品)
	type ProductInfoWithCategory struct {
		Name        string
		Description string
		Price       int64
		Category    string
	}
	products := []ProductInfoWithCategory{
		{"Joeby Tailored Trouser", "经典剪裁长裤，舒适面料，适合日常穿着。", 1799, "men"},        // product_1.jpg - 男士
		{"Denim Hooded", "牛仔连帽衫，时尚百搭，适合休闲场合。", 3000, "men"},                    // product_2.jpg - 男士
		{"Mint Maxi Dress", "薄荷绿长款连衣裙，优雅大方，适合多种场合。", 1799, "women"},          // product_3.jpg - 女士
		{"White Flounce Dress", "白色荷叶边连衣裙，清新甜美，展现女性魅力。", 1599, "women"},   // product_4.jpg - 女士
		{"Classic White Shirt", "经典白色衬衫，百搭单品，职场必备。", 1999, "men"},            // product_5.jpg - 男士
		{"Casual Denim Jacket", "休闲牛仔夹克，经典款式，四季可穿。", 2500, "men"},            // product_6.jpg - 男士
		{"Elegant Blouse", "优雅女士衬衫，精致剪裁，展现知性美。", 2200, "women"},              // product_7.jpg - 女士
		{"Stylish T-Shirt", "时尚T恤，舒适面料，休闲百搭。", 1299, "men"},                      // product_8.jpg - 男士
		{"Men's Belt", "男士皮带，真皮材质，经典设计。", 990, "accessories"},                    // product_9.jpg - 饰品
		{"Sport Hi Adidas", "阿迪达斯运动鞋，舒适透气，适合运动健身。", 2900, "accessories"},  // product_10.jpg - 饰品
		{"Leather Handbag", "真皮手提包，精致工艺，时尚优雅。", 3500, "women"},                  // product_11.jpg - 女士
		{"Designer Sunglasses", "设计师太阳镜，防紫外线，时尚潮流。", 1800, "accessories"},    // product_12.jpg - 饰品
	}

	// 设置秒杀时间：从现在开始，持续24小时
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	successCount := 0
	failCount := 0

	// 先获取所有商品列表
	resp, err := client.Get(adminURL + "/products")
	if err != nil {
		fmt.Printf("❌ 获取商品列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取商品列表响应失败: %v\n", err)
		return
	}

	var listResp struct {
		Code int `json:"code"`
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		fmt.Printf("❌ 解析商品列表失败: %v\n", err)
		return
	}

	if listResp.Code != 0 {
		fmt.Printf("❌ 获取商品列表失败: code=%d\n", listResp.Code)
		return
	}

	// 更新前12个商品
	for i := 1; i <= 12 && i <= len(listResp.Data); i++ {
		productID := listResp.Data[i-1].ID
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
			Category:     product.Category, // 设置分类
			Status:       1,               // 1:正常
			StartTime:    startTime,
			EndTime:      endTime,
		}

		jsonData, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("❌ 商品 %d: JSON 序列化失败: %v\n", i, err)
			failCount++
			continue
		}

		updateURL := fmt.Sprintf("%s/products/%d", adminURL, productID)
		httpReq, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ 商品 %d: 创建请求失败: %v\n", i, err)
			failCount++
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")

		updateResp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 请求失败: %v\n", i, err)
			failCount++
			continue
		}
		defer updateResp.Body.Close()

		updateBody, err := io.ReadAll(updateResp.Body)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 读取响应失败: %v\n", i, err)
			failCount++
			continue
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(updateBody, &apiResp); err != nil {
			fmt.Printf("❌ 商品 %d: JSON 解析失败: %v\n", i, err)
			fmt.Printf("   响应内容: %s\n", string(updateBody))
			failCount++
			continue
		}

		if apiResp.Code == 0 {
			fmt.Printf("✅ 商品 %d (%s) - ¥%.2f 更新成功\n", i, product.Name, float64(product.Price)/100)
			successCount++
		} else {
			fmt.Printf("❌ 商品 %d (%s) 更新失败: %s\n", i, product.Name, apiResp.Msg)
			failCount++
		}
	}

	fmt.Printf("\n📊 总结: 成功更新 %d 个, 失败 %d 个\n", successCount, failCount)
}

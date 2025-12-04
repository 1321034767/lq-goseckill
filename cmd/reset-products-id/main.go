package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/repository/mysql"
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
	client := &http.Client{}

	fmt.Println("🔄 重置商品ID为1-12（包含数据库AUTO_INCREMENT重置）...")
	fmt.Println("=" + strings.Repeat("=", 70))

	// 步骤1: 通过API删除所有商品
	fmt.Println("\n[1/4] 通过API删除所有商品...")
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
		Code int `json:"code"`
		Data []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
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

	fmt.Printf("📋 找到 %d 个商品\n", len(listResp.Data))

	deleteCount := 0
	for _, p := range listResp.Data {
		deleteURL := fmt.Sprintf("%s/products/%d", adminURL, p.ID)
		req, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 创建请求失败: %v\n", p.ID, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 商品 %d: 删除失败: %v\n", p.ID, err)
			continue
		}
		resp.Body.Close()
		deleteCount++
	}

	if deleteCount > 0 {
		fmt.Printf("✅ 已删除 %d 个商品\n", deleteCount)
	} else {
		fmt.Println("ℹ️  没有商品需要删除")
	}

	// 步骤2: 重置数据库AUTO_INCREMENT
	fmt.Println("\n[2/4] 重置数据库AUTO_INCREMENT...")
	cfg := config.DefaultConfig()
	db := mysql.Init(&cfg.MySQL)

	// 重置products表的AUTO_INCREMENT为1
	result := db.Exec("ALTER TABLE products AUTO_INCREMENT = 1")
	if result.Error != nil {
		fmt.Printf("❌ 重置AUTO_INCREMENT失败: %v\n", result.Error)
		return
	}
	fmt.Println("✅ AUTO_INCREMENT已重置为1")

	// 步骤3: 添加12个新商品
	fmt.Println("\n[3/4] 添加12个新商品（ID将从1开始）...")

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

	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	addSuccessCount := 0
	addFailCount := 0
	createdIDs := []int64{}

	for i := 1; i <= 12; i++ {
		product := products[i-1]

		stock := 50 + int64(i*5)
		seckillStock := 5 + int64(i/2)

		req := ProductRequest{
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Stock:        stock,
			SeckillStock: seckillStock,
			Category:     product.Category, // 设置分类
			Status:       1,
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
			addFailCount++
			continue
		}

		if apiResp.Code == 0 {
			var productData map[string]interface{}
			if data, ok := apiResp.Data.(map[string]interface{}); ok {
				productData = data
			} else if dataBytes, err := json.Marshal(apiResp.Data); err == nil {
				json.Unmarshal(dataBytes, &productData)
			}

			productID := int64(0)
			if id, ok := productData["id"].(float64); ok {
				productID = int64(id)
				createdIDs = append(createdIDs, productID)
			}

			fmt.Printf("✅ 商品 %d (%s) - ¥%.2f 添加成功 (ID: %d)\n", i, product.Name, float64(product.Price)/100, productID)
			addSuccessCount++
		} else {
			fmt.Printf("❌ 商品 %d (%s) 添加失败: %s\n", i, product.Name, apiResp.Msg)
			addFailCount++
		}
	}

	fmt.Printf("\n📊 添加总结: 成功 %d 个, 失败 %d 个\n", addSuccessCount, addFailCount)

	// 步骤4: 验证商品ID
	fmt.Println("\n[4/4] 验证商品ID...")
	if len(createdIDs) > 0 {
		minID := createdIDs[0]
		maxID := createdIDs[0]
		for _, id := range createdIDs {
			if id < minID {
				minID = id
			}
			if id > maxID {
				maxID = id
			}
		}

		fmt.Printf("📋 商品ID范围: %d - %d\n", minID, maxID)

		if minID == 1 && maxID == 12 {
			fmt.Println("✅ 验证成功！商品ID确实是 1-12")
			fmt.Println("\n🎉 完成！商品ID现在是 1-12")
			fmt.Println("   商品ID 1 → 图片 product_1.jpg")
			fmt.Println("   商品ID 2 → 图片 product_2.jpg")
			fmt.Println("   ...")
			fmt.Println("   商品ID 12 → 图片 product_12.jpg")
		} else {
			fmt.Printf("⚠️  警告：商品ID范围是 %d-%d，不是期望的 1-12\n", minID, maxID)
			fmt.Println("   请检查数据库AUTO_INCREMENT设置")
		}
	} else {
		fmt.Println("⚠️  无法验证：没有成功创建的商品")
	}
}

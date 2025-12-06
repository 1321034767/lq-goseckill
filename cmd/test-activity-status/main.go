package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/seckill_activity"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("测试活动状态自动更新功能")
	fmt.Println("==========================================")

	cfg := config.DefaultConfig()
	db := mysql.Init(&cfg.MySQL)
	activityRepo := mysql.NewSeckillActivityRepository(db)
	productRepo := mysql.NewProductRepository(db)
	activitySvc := service.NewSeckillActivityService(activityRepo, productRepo)

	ctx := context.Background()

	// 1. 获取所有活动
	fmt.Println("\n1. 获取所有活动（检查过期）...")
	activities, err := activityRepo.ListAll(ctx)
	if err != nil {
		fmt.Printf("❌ 获取活动列表失败: %v\n", err)
		return
	}

	if len(activities) == 0 {
		fmt.Println("⚠️  没有活动，请先创建活动")
		return
	}

	fmt.Printf("找到 %d 个活动\n", len(activities))
	now := time.Now()

	// 2. 显示活动状态（更新前）
	fmt.Println("\n2. 更新前的活动状态：")
	for _, a := range activities {
		statusText := getStatusText(a.Status)
		isExpired := now.After(a.EndTime)
		fmt.Printf("  活动ID: %d, 名称: %s, 状态: %s, 结束时间: %s, 是否已过期: %v\n",
			a.ID, a.Name, statusText, a.EndTime.Format("2006-01-02 15:04:05"), isExpired)
	}

	// 3. 执行过期检查
	fmt.Println("\n3. 执行过期检查...")
	if err := activitySvc.CheckAndUpdateExpiredActivities(ctx); err != nil {
		fmt.Printf("❌ 过期检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ 过期检查完成")

	// 4. 重新获取活动，显示更新后的状态
	fmt.Println("\n4. 更新后的活动状态：")
	activities, err = activityRepo.ListAll(ctx)
	if err != nil {
		fmt.Printf("❌ 重新获取活动列表失败: %v\n", err)
		return
	}

	for _, a := range activities {
		statusText := getStatusText(a.Status)
		isExpired := now.After(a.EndTime)
		shouldBeEnded := isExpired && a.Status != 2
		marker := ""
		if shouldBeEnded {
			marker = " ❌ 应该已结束但状态不正确"
		} else if isExpired && a.Status == 2 {
			marker = " ✅ 已正确更新为已结束"
		}
		fmt.Printf("  活动ID: %d, 名称: %s, 状态: %s, 结束时间: %s, 是否已过期: %v%s\n",
			a.ID, a.Name, statusText, a.EndTime.Format("2006-01-02 15:04:05"), isExpired, marker)
	}

	// 5. 检查商品状态
	fmt.Println("\n5. 检查关联商品状态...")
	for _, a := range activities {
		if now.After(a.EndTime) && a.Status == 2 {
			products, err := activityRepo.GetProductsByActivity(ctx, a.ID)
			if err != nil {
				continue
			}
			for _, ap := range products {
				p, err := productRepo.GetByID(ctx, ap.ProductID)
				if err != nil {
					continue
				}
				productStatusText := getProductStatusText(p.Status)
				marker := ""
				if p.Status == 2 {
					marker = " ❌ 商品状态仍为秒杀中，应该已恢复"
				} else if p.Status == 1 && p.SeckillStock == 0 {
					marker = " ✅ 商品状态已恢复为正常"
				}
				fmt.Printf("  商品ID: %d, 名称: %s, 状态: %s, 秒杀库存: %d%s\n",
					p.ID, p.Name, productStatusText, p.SeckillStock, marker)
			}
		}
	}

	// 6. 测试API接口
	fmt.Println("\n6. 测试后台管理API接口...")
	testAdminAPI()

	fmt.Println("\n==========================================")
	fmt.Println("测试完成！")
	fmt.Println("==========================================")
}

func getStatusText(status int) string {
	statusMap := map[int]string{
		0: "未开始",
		1: "进行中",
		2: "已结束",
		3: "已取消",
	}
	return statusMap[status]
}

func getProductStatusText(status int) string {
	statusMap := map[int]string{
		0: "下线",
		1: "正常",
		2: "秒杀中",
	}
	return statusMap[status]
}

func testAdminAPI() {
	// 测试活动列表API
	fmt.Println("\n  测试 GET /api/seckill-activities...")
	resp, err := http.Get("http://localhost:8081/api/seckill-activities")
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("  ❌ 读取响应失败: %v\n", err)
		return
	}

	var result struct {
		Code int                              `json:"code"`
		Data []*seckill_activity.SeckillActivity `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("  ❌ JSON解析失败: %v\n", err)
		fmt.Printf("  响应内容: %s\n", string(body))
		return
	}

	if result.Code != 0 {
		fmt.Printf("  ❌ API返回错误: code=%d\n", result.Code)
		return
	}

	now := time.Now()
	fmt.Printf("  ✅ API返回 %d 个活动\n", len(result.Data))
	for _, a := range result.Data {
		isExpired := now.After(a.EndTime)
		marker := ""
		if isExpired && a.Status == 2 {
			marker = " ✅ 已正确更新"
		} else if isExpired && a.Status != 2 {
			marker = " ❌ 应该已结束"
		}
		fmt.Printf("    活动ID: %d, 状态: %s, 结束时间: %s%s\n",
			a.ID, getStatusText(a.Status), a.EndTime.Format("2006-01-02 15:04:05"), marker)
	}

	// 测试商品列表API
	fmt.Println("\n  测试 GET /api/products...")
	resp2, err := http.Get("http://localhost:8081/api/products")
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		fmt.Printf("  ❌ 读取响应失败: %v\n", err)
		return
	}

	var result2 struct {
		Code int    `json:"code"`
		Data []struct {
			ID           int64 `json:"ID"`
			Name         string `json:"Name"`
			Status       int    `json:"Status"`
			SeckillStock int64 `json:"SeckillStock"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body2, &result2); err != nil {
		fmt.Printf("  ❌ JSON解析失败: %v\n", err)
		fmt.Printf("  响应内容: %s\n", string(body2))
		return
	}

	if result2.Code != 0 {
		fmt.Printf("  ❌ API返回错误: code=%d\n", result2.Code)
		return
	}

	fmt.Printf("  ✅ API返回 %d 个商品\n", len(result2.Data))
	seckillProducts := 0
	for _, p := range result2.Data {
		if p.Status == 2 {
			seckillProducts++
			fmt.Printf("    商品ID: %d, 名称: %s, 状态: %s, 秒杀库存: %d\n",
				p.ID, p.Name, getProductStatusText(p.Status), p.SeckillStock)
		}
	}
	if seckillProducts == 0 {
		fmt.Println("    ✅ 没有处于秒杀状态的商品（符合预期）")
	}
}

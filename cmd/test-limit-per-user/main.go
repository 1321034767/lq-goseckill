package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL  = "http://localhost:8080"
	adminURL = "http://localhost:8081"
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("    限购数量功能测试")
	fmt.Println("==========================================")
	fmt.Println()

	// 1. 注册用户
	fmt.Println("1. 注册用户...")
	err := register("testuser", "testpass")
	if err != nil {
		fmt.Printf("   注册失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 注册成功")
	}

	// 2. 登录获取token
	fmt.Println("\n2. 登录获取token...")
	token, err := login("testuser", "testpass")
	if err != nil {
		fmt.Printf("   ❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Token: %s...\n", token[:20])

	// 3. 创建秒杀活动（带限购数量）
	fmt.Println("\n3. 创建秒杀活动（限购数量=2）...")
	now := time.Now()
	startTime := now.Add(1 * time.Hour).Format("2006-01-02T15:04:05")
	endTime := now.Add(2 * time.Hour).Format("2006-01-02T15:04:05")

	activityData := map[string]interface{}{
		"name":           "限购测试活动",
		"description":    "测试每人限购2件",
		"start_time":     startTime,
		"end_time":       endTime,
		"discount":       0.8,
		"limit_per_user": 2,
		"product_ids":     []int64{1},
		"product_stocks": map[int64]int64{1: 100},
	}

	activityResp, err := httpPost(adminURL+"/api/seckill-activities", activityData, "")
	if err != nil {
		fmt.Printf("   ❌ 创建活动失败: %v\n", err)
		return
	}
	activityID, ok := activityResp["id"].(float64)
	if !ok {
		fmt.Printf("   ❌ 活动ID格式错误: %v\n", activityResp)
		return
	}
	fmt.Printf("   ✅ 活动创建成功，ID: %.0f\n", activityID)

	// 4. 启动活动
	fmt.Println("\n4. 启动活动...")
	startResp, err := httpPost(fmt.Sprintf("%s/api/seckill-activities/%.0f/start", adminURL, activityID), nil, "")
	if err != nil {
		fmt.Printf("   ❌ 启动活动失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 活动启动成功: %v\n", startResp)
	}

	// 5. 验证限购数量显示
	fmt.Println("\n5. 验证限购数量显示...")
	detailResp, err := httpGet(fmt.Sprintf("%s/api/seckill-activities/%.0f", adminURL, activityID), "")
	if err != nil {
		fmt.Printf("   ❌ 获取活动详情失败: %v\n", err)
	} else {
		activity, ok := detailResp["activity"].(map[string]interface{})
		if ok {
			limitPerUser := activity["limit_per_user"]
			fmt.Printf("   ✅ 限购数量: %v\n", limitPerUser)
			if limitPerUser == float64(2) {
				fmt.Println("   ✅ 限购数量设置正确")
			} else {
				fmt.Printf("   ❌ 限购数量错误，期望2，实际%v\n", limitPerUser)
			}
		}
	}

	// 6. 测试时间精确到秒
	fmt.Println("\n6. 验证时间精确到秒...")
	if activity, ok := detailResp["activity"].(map[string]interface{}); ok {
		startTimeStr := activity["start_time"].(string)
		endTimeStr := activity["end_time"].(string)
		fmt.Printf("   开始时间: %s\n", startTimeStr)
		fmt.Printf("   结束时间: %s\n", endTimeStr)
		
		// 检查时间格式是否包含秒
		if len(startTimeStr) >= 19 && startTimeStr[16:17] == ":" {
			fmt.Println("   ✅ 时间格式包含秒")
		} else {
			fmt.Println("   ❌ 时间格式不包含秒")
		}
	}

	fmt.Println("\n==========================================")
	fmt.Println("测试完成！")
	fmt.Println("==========================================")
	fmt.Println("\n提示：")
	fmt.Println("1. 请在后台管理界面查看活动列表，确认限购数量列显示正确")
	fmt.Println("2. 请在商品详情页查看限购数量显示")
	fmt.Println("3. 秒杀逻辑中的限购检查需要进一步实现")
}

func register(username, password string) error {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("%v", result["msg"])
	}
	return nil
}

func login(username, password string) (string, error) {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("登录失败: %s", result.Msg)
	}

	return result.Data.Token, nil
}

func httpGet(url, token string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 响应: %s", err, string(bodyBytes))
	}

	if code, ok := result["code"].(float64); ok && code != 0 {
		return nil, fmt.Errorf("API错误: %v", result["msg"])
	}

	return result, nil
}

func httpPost(url string, body interface{}, token string) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 响应: %s", err, string(bodyBytes))
	}

	if code, ok := result["code"].(float64); ok && code != 0 {
		return nil, fmt.Errorf("API错误: %v", result["msg"])
	}

	return result, nil
}

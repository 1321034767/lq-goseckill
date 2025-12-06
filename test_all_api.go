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
	fmt.Println("    完整API测试")
	fmt.Println("==========================================")
	fmt.Println()

	// 1. 注册用户
	fmt.Println("1. 注册用户...")
	registerResp, err := httpPost(baseURL+"/api/register", map[string]string{
		"username": "testuser",
		"password": "testpass",
	})
	if err != nil {
		fmt.Printf("   注册失败: %v\n", err)
	} else {
		fmt.Printf("   注册成功: %v\n", registerResp)
	}

	// 2. 登录获取token
	fmt.Println("\n2. 登录获取token...")
	loginResp, err := httpPost(baseURL+"/api/login", map[string]string{
		"username": "testuser",
		"password": "testpass",
	})
	if err != nil {
		fmt.Printf("   登录失败: %v\n", err)
		return
	}
	tokenData, ok := loginResp["data"].(map[string]interface{})
	if !ok {
		fmt.Printf("   登录响应格式错误: %v\n", loginResp)
		return
	}
	token, _ := tokenData["token"].(string)
	fmt.Printf("   Token: %s\n", token)

	// 3. 测试 /api/products/1/seckill-stock
	fmt.Println("\n3. 测试 /api/products/1/seckill-stock...")
	stockResp, err := httpGet(baseURL+"/api/products/1/seckill-stock", "")
	if err != nil {
		fmt.Printf("   失败: %v\n", err)
	} else {
		fmt.Printf("   成功: %v\n", stockResp)
	}

	// 4. 测试 /api/seckill/1/result
	fmt.Println("\n4. 测试 /api/seckill/1/result...")
	resultResp, err := httpGet(baseURL+"/api/seckill/1/result", token)
	if err != nil {
		fmt.Printf("   失败: %v\n", err)
	} else {
		fmt.Printf("   成功: %v\n", resultResp)
	}

	// 5. 测试 /api/monitor/stats (admin)
	fmt.Println("\n5. 测试 /api/monitor/stats (admin)...")
	monitorResp, err := httpGet(adminURL+"/api/monitor/stats", "")
	if err != nil {
		fmt.Printf("   失败: %v\n", err)
	} else {
		fmt.Printf("   成功: %v\n", monitorResp)
	}

	// 6. 测试限流
	fmt.Println("\n6. 测试限流功能...")
	pathResp, err := httpGet(baseURL+"/api/seckill/1/path", token)
	if err != nil {
		fmt.Printf("   获取路径失败: %v\n", err)
	} else {
		pathData, ok := pathResp["data"].(map[string]interface{})
		if ok {
			path, _ := pathData["path"].(string)
			fmt.Printf("   路径: %s\n", path)
			fmt.Println("   发送30个快速请求...")
			rateLimitCount := 0
			successCount := 0
			for i := 0; i < 30; i++ {
				resp, err := httpPost(baseURL+fmt.Sprintf("/api/seckill/1/%s", path), nil, token)
				if err != nil {
					if err.Error() == "429" {
						rateLimitCount++
					}
					continue
				}
				code, ok := resp["code"].(float64)
				if ok {
					if code == 429 {
						rateLimitCount++
					} else if code == 0 {
						successCount++
					}
				}
			}
			fmt.Printf("   成功: %d, 限流: %d\n", successCount, rateLimitCount)
		}
	}

	fmt.Println("\n==========================================")
	fmt.Println("测试完成！")
	fmt.Println("==========================================")
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

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 响应: %s", err, string(bodyBytes))
	}

	return result, nil
}

func httpPost(url string, body interface{}, token ...string) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", token[0])
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

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 响应: %s", err, string(bodyBytes))
	}

	return result, nil
}

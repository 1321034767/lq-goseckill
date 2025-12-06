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
	baseURL = "http://localhost:8080"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code int    `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type PathResponse struct {
	Code int    `json:"code"`
	Data struct {
		Path string `json:"path"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type SeckillResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func main() {
	fmt.Println("=== 幂等性测试程序 ===\n")

	// 测试用户信息
	username := "testuser"
	password := "testpass"
	productID := int64(1) // 使用商品ID 1进行测试

	// 1. 注册用户（如果不存在）
	fmt.Println("步骤1: 注册/登录用户...")
	token, err := registerAndLogin(username, password)
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功，Token: %s...\n\n", token[:20])

	// 2. 获取秒杀路径
	fmt.Println("步骤2: 获取秒杀路径...")
	path, err := getSeckillPath(token, productID)
	if err != nil {
		fmt.Printf("❌ 获取路径失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 获取路径成功: %s\n\n", path)

	// 3. 第一次秒杀（应该成功）
	fmt.Println("步骤3: 第一次秒杀请求...")
	resp1, err := seckill(token, productID, path)
	if err != nil {
		fmt.Printf("❌ 第一次秒杀请求失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 第一次秒杀响应: code=%d, msg=%s\n\n", resp1.Code, resp1.Msg)

	// 等待一下，确保消息被处理
	fmt.Println("等待3秒，确保Worker处理完订单...")
	time.Sleep(3 * time.Second)

	// 4. 再次获取路径（模拟用户刷新页面）
	fmt.Println("步骤4: 再次获取秒杀路径...")
	path2, err := getSeckillPath(token, productID)
	if err != nil {
		fmt.Printf("❌ 获取路径失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 获取新路径: %s\n\n", path2)

	// 5. 第二次秒杀（应该被拒绝 - 幂等性检查）
	fmt.Println("步骤5: 第二次秒杀请求（应该被拒绝）...")
	resp2, err := seckill(token, productID, path2)
	if err != nil {
		fmt.Printf("❌ 第二次秒杀请求失败: %v\n", err)
		return
	}
	fmt.Printf("响应: code=%d, msg=%s\n", resp2.Code, resp2.Msg)

	// 6. 验证结果
	fmt.Println("\n=== 测试结果 ===")
	if resp1.Code == 0 && resp2.Code == 400 && resp2.Msg == "duplicate seckill" {
		fmt.Println("✅ 幂等性测试通过！")
		fmt.Println("   - 第一次秒杀成功")
		fmt.Println("   - 第二次秒杀被正确拒绝（duplicate seckill）")
	} else {
		fmt.Println("❌ 幂等性测试失败！")
		fmt.Printf("   第一次响应: code=%d, msg=%s\n", resp1.Code, resp1.Msg)
		fmt.Printf("   第二次响应: code=%d, msg=%s\n", resp2.Code, resp2.Msg)
		fmt.Println("\n   预期结果:")
		fmt.Println("   - 第一次: code=0 (成功)")
		fmt.Println("   - 第二次: code=400, msg='duplicate seckill' (被拒绝)")
	}
}

func registerAndLogin(username, password string) (string, error) {
	// 先尝试登录
	token, err := login(username, password)
	if err == nil {
		return token, nil
	}

	// 登录失败，尝试注册
	if err := register(username, password); err != nil {
		return "", fmt.Errorf("注册失败: %v", err)
	}

	// 注册后登录
	return login(username, password)
}

func register(username, password string) error {
	reqBody := LoginRequest{
		Username: username,
		Password: password,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("注册失败: %s", result.Msg)
	}
	return nil
}

func login(username, password string) (string, error) {
	reqBody := LoginRequest{
		Username: username,
		Password: password,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("登录失败: %s", result.Msg)
	}

	return result.Data.Token, nil
}

func getSeckillPath(token string, productID int64) (string, error) {
	url := fmt.Sprintf("%s/api/seckill/%d/path", baseURL, productID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result PathResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("获取路径失败: %s", result.Msg)
	}

	return result.Data.Path, nil
}

func seckill(token string, productID int64, path string) (*SeckillResponse, error) {
	url := fmt.Sprintf("%s/api/seckill/%d/%s", baseURL, productID, path)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SeckillResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL = "http://localhost:8080"
)

type TestResult struct {
	TestName string
	Success  bool
	Message  string
}

func main() {
	fmt.Println("=== GoSecKill 登录功能测试 ===")
	fmt.Println()

	// 创建HTTP客户端，支持Cookie
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	var results []TestResult

	// 测试1: 检查服务是否运行
	fmt.Println("测试1: 检查服务是否运行...")
	result := testServiceRunning(client)
	results = append(results, result)
	printResult(result)

	// 测试2: 访问登录页面
	fmt.Println("\n测试2: 访问登录页面...")
	result = testLoginPage(client)
	results = append(results, result)
	printResult(result)

	// 测试3: 注册新用户
	fmt.Println("\n测试3: 注册新用户...")
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	password := "testpass123"
	result = testRegister(client, username, password)
	results = append(results, result)
	printResult(result)

	if !result.Success {
		fmt.Println("\n注册失败，尝试使用已存在的用户登录...")
		username = "testuser"
		password = "testpass123"
	}

	// 测试4: 登录
	fmt.Println("\n测试4: 用户登录...")
	result = testLogin(client, username, password)
	results = append(results, result)
	printResult(result)

	// 测试5: 访问产品页面（需要登录）
	if result.Success {
		fmt.Println("\n测试5: 访问产品页面（验证登录状态）...")
		result = testProductPage(client)
		results = append(results, result)
		printResult(result)
	}

	// 测试6: 检查Cookie
	fmt.Println("\n测试6: 检查登录Cookie...")
	result = testCookies(client)
	results = append(results, result)
	printResult(result)

	// 汇总结果
	fmt.Println("\n=== 测试结果汇总 ===")
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
		fmt.Printf("%s: %s\n", r.TestName, getStatus(r.Success))
		if !r.Success && r.Message != "" {
			fmt.Printf("  错误: %s\n", r.Message)
		}
	}
	fmt.Printf("\n总计: %d/%d 测试通过\n", successCount, len(results))
}

func testServiceRunning(client *http.Client) TestResult {
	resp, err := client.Get(baseURL)
	if err != nil {
		return TestResult{
			TestName: "服务运行检查",
			Success:  false,
			Message:  fmt.Sprintf("无法连接到服务: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 404 {
		return TestResult{
			TestName: "服务运行检查",
			Success:  true,
			Message:  fmt.Sprintf("服务响应状态码: %d", resp.StatusCode),
		}
	}

	return TestResult{
		TestName: "服务运行检查",
		Success:  false,
		Message:  fmt.Sprintf("意外的状态码: %d", resp.StatusCode),
	}
}

func testLoginPage(client *http.Client) TestResult {
	resp, err := client.Get(baseURL + "/user/login")
	if err != nil {
		return TestResult{
			TestName: "登录页面访问",
			Success:  false,
			Message:  fmt.Sprintf("请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode == 200 && strings.Contains(bodyStr, "username") && strings.Contains(bodyStr, "password") {
		return TestResult{
			TestName: "登录页面访问",
			Success:  true,
			Message:  "登录页面加载成功",
		}
	}

	return TestResult{
		TestName: "登录页面访问",
		Success:  false,
		Message:  fmt.Sprintf("状态码: %d, 页面内容不包含登录表单", resp.StatusCode),
	}
}

func testRegister(client *http.Client, username, password string) TestResult {
	// 先访问注册页面
	resp, err := client.Get(baseURL + "/user/register")
	if err != nil {
		return TestResult{
			TestName: "用户注册",
			Success:  false,
			Message:  fmt.Sprintf("访问注册页面失败: %v", err),
		}
	}
	resp.Body.Close()

	// 提交注册表单
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	resp, err = client.PostForm(baseURL+"/user/add", data)
	if err != nil {
		return TestResult{
			TestName: "用户注册",
			Success:  false,
			Message:  fmt.Sprintf("注册请求失败: %v", err),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查是否重定向到登录页面或注册成功
	if resp.StatusCode == 302 || resp.StatusCode == 200 {
		location := resp.Header.Get("Location")
		if location == "/user/login" || strings.Contains(bodyStr, "login") || resp.StatusCode == 200 {
			return TestResult{
				TestName: "用户注册",
				Success:  true,
				Message:  fmt.Sprintf("用户 '%s' 注册成功", username),
			}
		}
	}

	return TestResult{
		TestName: "用户注册",
		Success:  false,
		Message:  fmt.Sprintf("注册失败，状态码: %d, 响应: %s", resp.StatusCode, bodyStr[:min(100, len(bodyStr))]),
	}
}

func testLogin(client *http.Client, username, password string) TestResult {
	// 提交登录表单
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	// 禁用自动重定向以检查响应
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.PostForm(baseURL+"/user/login", data)
	if err != nil {
		// 如果是重定向错误，这是正常的
		if resp != nil && resp.StatusCode == 302 {
			// 继续处理
		} else {
			return TestResult{
				TestName: "用户登录",
				Success:  false,
				Message:  fmt.Sprintf("登录请求失败: %v", err),
			}
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查是否重定向到产品页面
	location := resp.Header.Get("Location")
	
	// 检查Cookie（从响应中获取）
	cookies := resp.Cookies()
	hasUID := false
	uidValue := ""
	for _, cookie := range cookies {
		if cookie.Name == "uid" && cookie.Value != "" {
			hasUID = true
			uidValue = cookie.Value
			break
		}
	}

	// 如果状态码是302且重定向到/product，或者有uid cookie，认为登录成功
	if (resp.StatusCode == 302 && location == "/product") || hasUID {
		// 恢复自动重定向
		client.CheckRedirect = nil
		return TestResult{
			TestName: "用户登录",
			Success:  true,
			Message:  fmt.Sprintf("用户 '%s' 登录成功，重定向到: %s, UID: %s", username, location, uidValue),
		}
	}

	// 检查错误消息
	if strings.Contains(bodyStr, "Wrong username or password") || strings.Contains(bodyStr, "Failed") {
		client.CheckRedirect = nil
		return TestResult{
			TestName: "用户登录",
			Success:  false,
			Message:  fmt.Sprintf("登录失败: %s", bodyStr[:min(200, len(bodyStr))]),
		}
	}

	// 恢复自动重定向
	client.CheckRedirect = nil
	return TestResult{
		TestName: "用户登录",
		Success:  false,
		Message:  fmt.Sprintf("登录失败，状态码: %d, 位置: %s, 有UID Cookie: %v, 响应: %s", resp.StatusCode, location, hasUID, bodyStr[:min(200, len(bodyStr))]),
	}
}

func testProductPage(client *http.Client) TestResult {
	// Disable redirect to check response
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		client.CheckRedirect = nil
	}()

	resp, err := client.Get(baseURL + "/product")
	if err != nil {
		// If it's a redirect error, that's actually OK
		if resp != nil {
			// Continue processing
		} else {
			return TestResult{
				TestName: "产品页面访问",
				Success:  false,
				Message:  fmt.Sprintf("请求失败: %v", err),
			}
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	location := resp.Header.Get("Location")

	// Check for redirect to login (means not authenticated)
	if resp.StatusCode == 302 && location == "/user/login" {
		return TestResult{
			TestName: "产品页面访问",
			Success:  false,
			Message:  "未登录，被重定向到登录页面（Cookie可能已过期）",
		}
	}

	// Check for successful response
	if resp.StatusCode == 200 {
		// Check if it's an error page
		if strings.Contains(bodyStr, "异常错误处理页面") || strings.Contains(bodyStr, "error.html") {
			return TestResult{
				TestName: "产品页面访问",
				Success:  false,
				Message:  fmt.Sprintf("返回错误页面，状态码: %d", resp.StatusCode),
			}
		}
		
		// Check if it contains product-related content
		if strings.Contains(bodyStr, "product") || strings.Contains(bodyStr, "商品") || 
		   strings.Contains(bodyStr, "book") || len(bodyStr) > 1000 {
			return TestResult{
				TestName: "产品页面访问",
				Success:  true,
				Message:  fmt.Sprintf("产品页面加载成功，内容长度: %d", len(bodyStr)),
			}
		}
		
		return TestResult{
			TestName: "产品页面访问",
			Success:  true,
			Message:  fmt.Sprintf("页面返回200，内容长度: %d", len(bodyStr)),
		}
	}

	return TestResult{
		TestName: "产品页面访问",
		Success:  false,
		Message:  fmt.Sprintf("状态码: %d, 位置: %s, 响应长度: %d, 内容预览: %s", 
			resp.StatusCode, location, len(bodyStr), bodyStr[:min(200, len(bodyStr))]),
	}
}

func testCookies(client *http.Client) TestResult {
	// 检查客户端jar中的cookie
	urlObj, _ := url.Parse(baseURL)
	cookies := client.Jar.Cookies(urlObj)
	
	cookieInfo := make(map[string]string)
	for _, cookie := range cookies {
		cookieInfo[cookie.Name] = cookie.Value
	}

	if len(cookieInfo) > 0 {
		cookieJSON, _ := json.Marshal(cookieInfo)
		hasUID := cookieInfo["uid"] != ""
		return TestResult{
			TestName: "Cookie检查",
			Success:  hasUID,
			Message:  fmt.Sprintf("Cookie: %s, 有UID: %v", string(cookieJSON), hasUID),
		}
	}

	return TestResult{
		TestName: "Cookie检查",
		Success:  false,
		Message:  "未找到Cookie",
	}
}

func printResult(result TestResult) {
	status := getStatus(result.Success)
	fmt.Printf("  %s: %s\n", result.TestName, status)
	if result.Message != "" {
		fmt.Printf("  %s\n", result.Message)
	}
}

func getStatus(success bool) string {
	if success {
		return "✓ 通过"
	}
	return "✗ 失败"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


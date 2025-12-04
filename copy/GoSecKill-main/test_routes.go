package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	fmt.Println("=== 路由调试测试 ===")
	fmt.Println()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	// 先登录获取Cookie
	fmt.Println("步骤1: 登录获取Cookie...")
	loginData := "username=testuser&password=testpass123"
	resp, err := client.Post(baseURL+"/user/login", "application/x-www-form-urlencoded", strings.NewReader(loginData))
	if err == nil {
		resp.Body.Close()
	}

	// 测试各种路由
	routes := []struct {
		method string
		path   string
		name   string
	}{
		{"GET", "/", "根路径"},
		{"GET", "/user/login", "登录页面"},
		{"GET", "/user/register", "注册页面"},
		{"GET", "/product", "产品页面"},
		{"GET", "/product/detail", "产品详情"},
		{"GET", "/product/order", "产品订单"},
		{"GET", "/assets/css/site.css", "静态资源"},
	}

	fmt.Println("\n步骤2: 测试各个路由...")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-20s %-8s %-30s %-10s %s\n", "路由名称", "方法", "路径", "状态码", "说明")
	fmt.Println(strings.Repeat("-", 80))

	for _, route := range routes {
		var req *http.Request
		if route.method == "GET" {
			req, _ = http.NewRequest("GET", baseURL+route.path, nil)
		} else {
			req, _ = http.NewRequest(route.method, baseURL+route.path, nil)
		}

		resp, err := client.Do(req)
		statusCode := 0
		note := ""

		if err != nil {
			note = fmt.Sprintf("错误: %v", err)
		} else {
			statusCode = resp.StatusCode
			location := resp.Header.Get("Location")
			
			if resp.StatusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)
				if strings.Contains(bodyStr, "异常错误处理页面") {
					note = "返回错误处理页面"
				} else {
					note = fmt.Sprintf("内容长度: %d", len(bodyStr))
				}
			} else if resp.StatusCode == 302 {
				note = fmt.Sprintf("重定向到: %s", location)
			} else if resp.StatusCode == 404 {
				note = "路由不存在"
			} else {
				note = fmt.Sprintf("状态码: %d", resp.StatusCode)
			}
			resp.Body.Close()
		}

		fmt.Printf("%-20s %-8s %-30s %-10d %s\n", route.name, route.method, route.path, statusCode, note)
	}

	fmt.Println(strings.Repeat("-", 80))
	
	// 详细测试产品路由
	fmt.Println("\n步骤3: 详细测试产品路由...")
	testProductRoute(client)
}

func testProductRoute(client *http.Client) {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		client.CheckRedirect = nil
	}()

	fmt.Println("\n测试 /product 路由:")
	fmt.Println(strings.Repeat("-", 80))

	// 测试1: 不带Cookie
	fmt.Println("测试1: 不带Cookie访问...")
	req1, _ := http.NewRequest("GET", baseURL+"/product", nil)
	resp1, _ := client.Do(req1)
	if resp1 != nil {
		fmt.Printf("  状态码: %d\n", resp1.StatusCode)
		fmt.Printf("  位置: %s\n", resp1.Header.Get("Location"))
		resp1.Body.Close()
	}

	// 测试2: 带Cookie
	fmt.Println("\n测试2: 带Cookie访问...")
	req2, _ := http.NewRequest("GET", baseURL+"/product", nil)
	req2.Header.Set("Cookie", "uid=1")
	resp2, _ := client.Do(req2)
	if resp2 != nil {
		fmt.Printf("  状态码: %d\n", resp2.StatusCode)
		fmt.Printf("  位置: %s\n", resp2.Header.Get("Location"))
		body, _ := io.ReadAll(resp2.Body)
		bodyStr := string(body)
		fmt.Printf("  响应长度: %d\n", len(bodyStr))
		fmt.Printf("  包含'product': %v\n", strings.Contains(strings.ToLower(bodyStr), "product"))
		fmt.Printf("  包含'错误': %v\n", strings.Contains(bodyStr, "错误") || strings.Contains(bodyStr, "error"))
		fmt.Printf("  内容预览: %s\n", bodyStr[:min(150, len(bodyStr))])
		resp2.Body.Close()
	}

	// 测试3: 检查Cookie
	fmt.Println("\n测试3: 检查客户端Cookie...")
	urlObj, _ := url.Parse(baseURL)
	cookies := client.Jar.Cookies(urlObj)
	if len(cookies) > 0 {
		fmt.Println("  找到Cookie:")
		for _, cookie := range cookies {
			fmt.Printf("    %s = %s\n", cookie.Name, cookie.Value)
		}
	} else {
		fmt.Println("  未找到Cookie")
	}

	fmt.Println(strings.Repeat("-", 80))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


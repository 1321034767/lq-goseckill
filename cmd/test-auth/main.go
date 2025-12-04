package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	base := os.Getenv("GOSECKILL_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 10 * time.Second,
	}

	// 1) 检查首页右上角用户按钮 href 是否为 javascript:void(0)
	checkUserMenu(client, base+"/")

	// 2) 打开登录页
	mustGet(client, base+"/login")

	username := fmt.Sprintf("tester_%d", time.Now().Unix())
	password := "Passw0rd!"

	registerForm := url.Values{"username": {username}, "password": {password}}
	mustPost(client, base+"/user/add", registerForm, http.StatusFound)

	loginForm := url.Values{"username": {username}, "password": {password}}
	mustPost(client, base+"/user/login", loginForm, http.StatusFound)

	fmt.Printf("✅ 测试成功：用户 %s 已完成注册并能登录，且首页用户菜单已按预期配置。\n", username)
}

// checkUserMenu 确认首页右上角用户按钮不会直接跳转到用户中心，而是纯菜单触发器。
func checkUserMenu(client *http.Client, target string) {
	resp, err := client.Get(target)
	if err != nil {
		panic(fmt.Sprintf("GET %s 失败: %v", target, err))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("读取 %s 响应失败: %v", target, err))
	}
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("GET %s 返回状态码 %d", target, resp.StatusCode))
	}
	html := string(body)
	if !strings.Contains(html, `id="top-bar__sign-in"`) {
		panic("首页未找到顶栏用户按钮 (id=\"top-bar__sign-in\")")
	}
	if !strings.Contains(html, `href="javascript:void(0)" class="top-bar__sign-in" id="top-bar__sign-in"`) {
		panic("顶栏用户按钮 href 不是 javascript:void(0)，可能会直接跳转到页面")
	}
	fmt.Println("检查首页用户菜单 OK：顶栏用户按钮为纯下拉触发，不会直接跳转。")
}

func mustGet(client *http.Client, target string) {
	resp, err := client.Get(target)
	if err != nil {
		panic(fmt.Sprintf("GET %s 失败: %v", target, err))
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("GET %s 返回状态码 %d", target, resp.StatusCode))
	}
	fmt.Printf("GET %s -> %d\n", target, resp.StatusCode)
}

func mustPost(client *http.Client, target string, form url.Values, wantStatus int) {
	resp, err := client.PostForm(target, form)
	if err != nil {
		panic(fmt.Sprintf("POST %s 失败: %v", target, err))
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != wantStatus {
		panic(fmt.Sprintf("POST %s 返回状态码 %d, 期望 %d", target, resp.StatusCode, wantStatus))
	}
	fmt.Printf("POST %s -> %d\n", target, resp.StatusCode)
}

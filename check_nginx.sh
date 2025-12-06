#!/bin/bash

# 用途：快速排查域名 -> Nginx -> 后端端口链路
# 使用方式：
#   chmod +x check_nginx.sh
#   ./check_nginx.sh [AKAINA_CN_DOMAIN] [AKAINA_SITE_DOMAIN] [EXPECTED_IP] [SKIP_EXTERNAL=true|false]

AKAINA_CN_DOMAIN="${1:-akaina.cn}"
AKAINA_SITE_DOMAIN="${2:-akaina.site}"
EXPECTED_IP="${3:-43.128.79.42}"
SKIP_EXTERNAL="${4:-false}"
NGINX_SITES_DIR="/etc/nginx/sites-enabled"

ok()   { echo -e "  ✅ $*"; }
warn() { echo -e "  ⚠️  $*"; }
err()  { echo -e "  ❌ $*"; }
section() { echo -e "\n==================== $* ===================="; }

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

check_dns() {
    local domain="$1"
    local expected="$2"
    section "DNS 解析 - $domain"
    if ! command_exists dig; then
        err "未安装 dig (dnsutils)，无法检测 DNS"
        return
    fi
    local result
    result="$(dig +short "$domain" A | tr '\n' ' ')"
    if [[ -z "$result" ]]; then
        err "$domain 无 A 记录"
    else
        echo "  解析结果：$result"
        if grep -qw "$expected" <<<"$result"; then
            ok "包含期望 IP $expected"
        else
            warn "期望 IP $expected 未出现在解析结果中"
        fi
    fi
}

check_nginx_config() {
    section "Nginx 配置"
    if ! command_exists nginx; then
        err "未安装 Nginx"
        return
    fi

    if nginx -t >/tmp/nginx_test.log 2>&1; then
        ok "nginx -t 语法检查通过"
    else
        err "nginx -t 语法错误："
        cat /tmp/nginx_test.log
    fi

    if [[ -d "$NGINX_SITES_DIR" ]]; then
        echo "  已启用站点："
        ls -l "$NGINX_SITES_DIR"
    else
        warn "$NGINX_SITES_DIR 不存在"
    fi
}

check_port() {
    local port="$1"
    local desc="$2"
    section "端口检测 - $desc ($port)"
    if command_exists ss; then
        if ss -tlnp | grep -q ":$port "; then
            ok "端口 $port 正在监听"
            ss -tlnp | grep ":$port "
        else
            err "端口 $port 未监听"
        fi
    elif command_exists netstat; then
        if netstat -tlnp | grep -q ":$port "; then
            ok "端口 $port 正在监听"
            netstat -tlnp | grep ":$port "
        else
            err "端口 $port 未监听"
        fi
    else
        err "缺少 ss/netstat，无法检查端口"
    fi
}

check_curl() {
    local host="$1"
    local path="${2:-/}"
    local target="${3:-http://127.0.0.1}"
    local desc="$4"
    local expect_status="${5:-200}"
    section "cURL 检查 - $desc"
    if ! command_exists curl; then
        err "未安装 curl"
        return
    fi
    local code
    code="$(curl -s -o /dev/null -w "%{http_code}" -H "Host: $host" "$target$path")"
    LAST_CODE="$code"
    if [[ -n "$code" ]]; then
        echo "  响应码：$code"
        if [[ "$code" == "000" ]]; then
            warn "外网请求未建立连接，可能为安全组/出口限制，可用 curl --resolve 强制回环，或传第四个参数 true 跳过外网检测"
        elif [[ "$code" == "$expect_status" || ( "$expect_status" == "200" && ( "$code" == "301" || "$code" == "302" ) ) ]]; then
            ok "cURL 请求成功 ($desc)"
        elif [[ "$host" == "$AKAINA_SITE_DOMAIN" && "$path" == "/api" && "$code" == "404" ]]; then
            warn "8081 已打通但 /api 返回 404，检查后端路由是否存在或路径是否正确"
        else
            warn "cURL 返回码 $code，需检查后端"
        fi
    else
        err "cURL 请求失败，检查网络或 Nginx"
    fi
}

main() {
    echo "==========================================="
    echo "Nginx / DNS / 端口 / cURL 综合巡检脚本"
    echo "检查时间：$(date '+%F %T')"
    echo "==========================================="

    check_dns "$AKAINA_CN_DOMAIN" "$EXPECTED_IP"
    check_dns "$AKAINA_SITE_DOMAIN" "$EXPECTED_IP"

    check_nginx_config

    check_port 80  "Nginx HTTP"
    check_port 5000 "akaina.cn upstream"
    check_port 8080 "akaina.site UI"
    check_port 8081 "akaina.site API"

    check_curl "$AKAINA_CN_DOMAIN" "/" "http://127.0.0.1" "内网代理 (akaina.cn -> 5000)"
    check_curl "$AKAINA_SITE_DOMAIN" "/" "http://127.0.0.1" "内网代理 (akaina.site -> 8080)"
    check_curl "$AKAINA_SITE_DOMAIN" "/api" "http://127.0.0.1" "内网代理 (akaina.site/api -> 8081)" "200"

    if [[ "$SKIP_EXTERNAL" != "true" ]]; then
        check_curl "$AKAINA_CN_DOMAIN" "/" "http://$AKAINA_CN_DOMAIN" "外网直连 (akaina.cn)"
        check_curl "$AKAINA_SITE_DOMAIN" "/" "http://$AKAINA_SITE_DOMAIN" "外网直连 (akaina.site)"
    else
        echo -e "\n已根据参数跳过外网连通性检查。"
    fi

    echo -e "\n检查看完，如发现 ❌ 或 ⚠️ 请针对对应环节排查。若需忽略外网检测，可执行：./check_nginx.sh $AKAINA_CN_DOMAIN $AKAINA_SITE_DOMAIN $EXPECTED_IP true"
}

main "$@"

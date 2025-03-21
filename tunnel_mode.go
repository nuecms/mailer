package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

// EnableTunnelMode 更改配置以支持 Cloudflare Tunnel 模式
func EnableTunnelMode() {
	log.Println("启用 Cloudflare Tunnel 模式")
	
	// 添加一个简单的端点，用于检测 Tunnel 是否正常工作
	http.HandleFunc("/tunnel-test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","message":"Tunnel is working"}`)
	})
	
	// 打印提醒信息
	log.Println("重要: 在 Tunnel 模式下，您需要:")
	log.Println("1. 配置 config.json，设置 security.allowLocalOnly: false")
	log.Println("2. 确保设置了强密码（用于SMTP认证）")
	log.Println("3. 在您的应用中使用完整的SMTP地址、端口和认证信息")
}

// IsTunnelRequest 检查请求是否来自 Cloudflare Tunnel
func IsTunnelRequest(addr net.Addr) bool {
	// 从地址获取IP
	ipStr, _, _ := net.SplitHostPort(addr.String())
	ip := net.ParseIP(ipStr)
	
	// Cloudflare Tunnel 请求通常来自本地回环地址
	if ip != nil && ip.IsLoopback() {
		return true
	}
	
	// 也可以检查 Cloudflare 的 IP 范围
	// 这里需要一个全面的 Cloudflare IP 范围列表
	
	return false
}

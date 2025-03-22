// This file is being deprecated and will be removed in future versions.
// Cloudflare Tunnel doesn't support port 25 by default, making it unsuitable for SMTP traffic.
// See: https://developers.cloudflare.com/cloudflare-one/faq/troubleshooting/#i-cannot-send-emails-on-port-25

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

// EnableTunnelMode is deprecated and will be removed in a future release
// Cloudflare WARP client blocks outgoing SMTP traffic on port 25 by default
func EnableTunnelMode() {
	log.Println("警告: Cloudflare Tunnel 模式已被废弃")
	log.Println("Cloudflare WARP 客户端默认会阻止端口 25 上的出站 SMTP 流量")
	log.Println("详情请参阅 Cloudflare 文档: https://developers.cloudflare.com/cloudflare-one/faq/troubleshooting/#i-cannot-send-emails-on-port-25")
	
	// Add an endpoint for consistency but discourage its use
	http.HandleFunc("/tunnel-deprecated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"deprecated","message":"Cloudflare Tunnel mode is deprecated"}`)
	})
}

// IsTunnelRequest is deprecated and will be removed in a future release
func IsTunnelRequest(addr net.Addr) bool {
	// For backward compatibility, just return false
	return false
}

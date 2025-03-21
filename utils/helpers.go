package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// IsLocalConnection 检查连接是否来自本地
func IsLocalConnection(addr net.Addr) bool {
	// 获取IP地址
	ipStr, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		log.Printf("无法解析地址 %v: %v", addr, err)
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Printf("无效的IP地址: %s", ipStr)
		return false
	}

	// 检查是否是本地IP (127.0.0.1, ::1)
	return ip.IsLoopback() || ipStr == "::1" || strings.HasPrefix(ipStr, "127.")
}

// ComputeCRAMMD5 计算CRAM-MD5摘要
func ComputeCRAMMD5(challenge, secret string) string {
	h := hmac.New(md5.New, []byte(secret))
	h.Write([]byte(challenge))
	digest := hex.EncodeToString(h.Sum(nil))
	return digest
}

// SummarizeRecipients 摘要展示收件人列表
func SummarizeRecipients(recipients []string) string {
	if len(recipients) <= 3 {
		return strings.Join(recipients, ", ")
	}
	return fmt.Sprintf("%s, ... (共 %d 位收件人)",
		strings.Join(recipients[:3], ", "), len(recipients))
}

// Min 返回两个整数中较小的一个
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SendAlert 发送告警
func SendAlert(message string) {
	log.Printf("告警: %s", message)
	// 在这里可以添加更多告警方式，如发送HTTP请求到监控系统
}

// CheckQueueBacklog 检查队列积压
func CheckQueueBacklog() {
	dir := "emails/failed"
	fileCnt, err := CountFilesInDir(dir)
	if err != nil {
		return // 目录不存在或无法访问
	}

	if fileCnt > 100 {
		SendAlert(fmt.Sprintf("邮件队列积压: %d 封邮件等待处理", fileCnt))
	}
}

// CountFilesInDir 计算目录中的文件数量
func CountFilesInDir(dir string) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	return len(files), nil
}

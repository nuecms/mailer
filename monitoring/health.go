package monitoring

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/nuecms/mailer/mail"
)

// SystemHealthCheck 检查系统状态
func SystemHealthCheck(metrics *Metrics) map[string]interface{} {
	result := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"details":   map[string]interface{}{},
	}

	// 检查磁盘空间
	var stat syscall.Statfs_t
	if err := syscall.Statfs(".", &stat); err == nil {
		available := stat.Bavail * uint64(stat.Bsize)
		result["details"].(map[string]interface{})["disk_space_available_mb"] = available / 1024 / 1024 // MB
	} else {
		result["details"].(map[string]interface{})["disk_space_error"] = err.Error()
	}

	// 检查队列大小
	emailsDir := "emails"
	if _, err := os.Stat(emailsDir); err == nil {
		files, err := os.ReadDir(emailsDir)
		if err == nil {
			result["details"].(map[string]interface{})["queued_emails"] = len(files)
		} else {
			result["details"].(map[string]interface{})["queued_emails_error"] = err.Error()
		}
	}

	// 检查失败邮件数量
	failedDir := "emails/failed"
	if _, err := os.Stat(failedDir); err == nil {
		files, err := os.ReadDir(failedDir)
		if err == nil {
			result["details"].(map[string]interface{})["failed_emails"] = len(files)
		} else {
			result["details"].(map[string]interface{})["failed_emails_error"] = err.Error()
		}
	}

	// 添加运行时统计
	if metrics != nil {
		metrics.Mu.Lock()
		result["details"].(map[string]interface{})["total_emails_processed"] = metrics.TotalEmails
		result["details"].(map[string]interface{})["success_rate"] = 0.0
		if metrics.TotalEmails > 0 {
			result["details"].(map[string]interface{})["success_rate"] =
				float64(metrics.SuccessfulEmails) / float64(metrics.TotalEmails) * 100.0
		}
		metrics.Mu.Unlock()
	}

	return result
}

// StartHealthCheckServer 启动健康检查HTTP服务
func StartHealthCheckServer(port int, metrics *Metrics) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 只允许本地连接
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip := net.ParseIP(host)
			if ip != nil && !ip.IsLoopback() {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		health := SystemHealthCheck(metrics)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 只允许本地连接
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip := net.ParseIP(host)
			if ip != nil && !ip.IsLoopback() {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		result := metrics.GetMetricsData()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/admin/retry-failed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 只允许本地连接
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip := net.ParseIP(host)
			if ip != nil && !ip.IsLoopback() {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		go mail.ProcessFailedEmails()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "已开始处理失败邮件",
		})
	})

	// 尝试不同的端口，如果主端口被占用
	tryPorts := []int{port, port + 1, port + 2, 8125, 8225, 8325}
	
	var err error
	var listener net.Listener
	var usedPort int
	
	// 尝试多个端口
	for _, p := range tryPorts {
		addr := fmt.Sprintf("127.0.0.1:%d", p)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			usedPort = p
			break
		}
		log.Printf("端口 %d 已被占用，尝试下一个端口", p)
	}
	
	if err != nil {
		log.Printf("无法启动健康检查服务，所有尝试的端口都被占用: %v", err)
		return
	}
	
	log.Printf("健康检查服务启动在 http://127.0.0.1:%d", usedPort)
	log.Printf("可用端点: /health, /metrics, /admin/retry-failed (POST)")
	
	server := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	
	if err := server.Serve(listener); err != nil {
		log.Printf("健康检查HTTP服务运行失败: %v", err)
	}
}

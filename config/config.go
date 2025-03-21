package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// Config 存储应用配置
type Config struct {
	SMTPHost string `json:"smtpHost"`
	SMTPPort int    `json:"smtpPort"`

	DefaultUsername string `json:"defaultUsername"`
	DefaultPassword string `json:"defaultPassword"`

	ForwardSMTP     bool   `json:"forwardSMTP"`
	ForwardHost     string `json:"forwardHost"`
	ForwardPort     int    `json:"forwardPort"`
	ForwardUsername string `json:"forwardUsername"`
	ForwardPassword string `json:"forwardPassword"`
	ForwardSSL      bool   `json:"forwardSSL"`

	// 新增配置选项
	BatchSize         int  `json:"batchSize"`
	BatchDelay        int  `json:"batchDelay"`
	EnableHealthCheck bool `json:"enableHealthCheck"`
	HealthCheckPort   int  `json:"healthCheckPort"`

	RateLimits struct {
		Enabled    bool `json:"enabled"`
		MaxPerHour int  `json:"maxPerHour"`
		MaxPerDay  int  `json:"maxPerDay"`
	} `json:"rateLimits"`

	Security struct {
		AllowLocalOnly bool `json:"allowLocalOnly"`
		LogAllEmails   bool `json:"logAllEmails"`
		RequireAuth    bool `json:"requireAuth"`
	} `json:"security"`
}

// Load 从指定路径加载配置
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.BatchSize <= 0 {
		config.BatchSize = 20
	}
	if config.BatchDelay <= 0 {
		config.BatchDelay = 1000
	}
	if config.HealthCheckPort <= 0 {
		config.HealthCheckPort = 8025
	}

	return config, nil
}

// CheckForwardingConfig 检查转发设置
func CheckForwardingConfig(config *Config) {
	if !config.ForwardSMTP {
		log.Printf("警告: 转发功能已禁用，邮件将只保存在本地")
		return
	}

	if config.ForwardHost == "" {
		log.Printf("警告: 未设置转发主机，邮件将只保存在本地")
		return
	}

	if config.ForwardHost == "smtp.gmail.com" && !strings.Contains(config.ForwardUsername, "@gmail.com") {
		log.Printf("警告: Gmail转发用户名应该是完整Gmail地址，当前: %s", config.ForwardUsername)
	}

	log.Printf("转发配置检查完成: 将使用 %s:%d 发送邮件", config.ForwardHost, config.ForwardPort)
}

// MaskPassword 隐藏密码，只显示前两位和后两位
func MaskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}

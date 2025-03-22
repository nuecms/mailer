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

	// 直接发送配置
	DirectDelivery *DirectDeliveryConfig `json:"directDelivery"`

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

	// DKIM 配置
	DKIM *DKIMConfig `json:"dkim"`
}

// DKIMConfig 存储DKIM签名配置
type DKIMConfig struct {
	Enabled         bool     `json:"enabled"`           // 是否启用DKIM签名
	Domain          string   `json:"domain"`            // DKIM域名，通常是发件人的域名
	Selector        string   `json:"selector"`          // DKIM选择器
	PrivateKeyPath  string   `json:"privateKeyPath"`    // DKIM私钥路径
	HeadersToSign   []string `json:"headersToSign"`     // 要签名的头部字段
	SignatureExpiry int64    `json:"signatureExpiry"`   // 签名过期时间（秒）
}

// DirectDeliveryConfig 存储直接发送邮件的配置
type DirectDeliveryConfig struct {
	Enabled            bool   `json:"enabled"`            // 是否启用直接发送
	EhloDomain         string `json:"ehloDomain"`         // 用于EHLO的域名
	InsecureSkipVerify bool   `json:"insecureSkipVerify"` // 是否跳过TLS验证
	RetryCount         int    `json:"retryCount"`         // 重试次数
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

// CheckAllConfig 检查所有配置
func CheckAllConfig(config *Config) {
	CheckForwardingConfig(config)
	CheckDirectDeliveryConfig(config)
	CheckDKIMConfig(config)
}

// CheckDirectDeliveryConfig 检查直接发送设置
func CheckDirectDeliveryConfig(config *Config) {
	if config.DirectDelivery != nil && config.DirectDelivery.Enabled {
		log.Printf("直接发送模式已启用，将尝试直接发送邮件到目标服务器")
		
		if config.DirectDelivery.EhloDomain != "" {
			log.Printf("将使用 %s 作为EHLO域名", config.DirectDelivery.EhloDomain)
		}
		
		if config.DirectDelivery.InsecureSkipVerify {
			log.Printf("警告: TLS验证已禁用，这可能降低安全性")
		}
		
		if config.DirectDelivery.RetryCount <= 0 {
			config.DirectDelivery.RetryCount = 3
			log.Printf("设置默认重试次数为 %d", config.DirectDelivery.RetryCount)
		}
	}
}

// CheckDKIMConfig 检查DKIM配置
func CheckDKIMConfig(config *Config) {
	if config.DKIM == nil || !config.DKIM.Enabled {
		return // DKIM 未启用，跳过
	}

	log.Printf("DKIM签名已启用")
	
	if config.DKIM.Domain == "" {
		log.Printf("警告: DKIM域名未设置，将使用默认用户名域名")
		if strings.Contains(config.DefaultUsername, "@") {
			parts := strings.Split(config.DefaultUsername, "@")
			config.DKIM.Domain = parts[1]
			log.Printf("设置DKIM域名为: %s", config.DKIM.Domain)
		} else {
			log.Printf("警告: 无法确定DKIM域名，DKIM签名可能无法正常工作")
		}
	}
	
	if config.DKIM.Selector == "" {
		config.DKIM.Selector = "mail" // 默认选择器
		log.Printf("DKIM选择器未设置，使用默认值: %s", config.DKIM.Selector)
	}
	
	if config.DKIM.PrivateKeyPath == "" {
		config.DKIM.PrivateKeyPath = fmt.Sprintf("keys/%s/%s.private", config.DKIM.Domain, config.DKIM.Selector)
		log.Printf("DKIM私钥路径未设置，使用默认路径: %s", config.DKIM.PrivateKeyPath)
	}
	
	// 检查私钥文件是否存在
	if _, err := os.Stat(config.DKIM.PrivateKeyPath); os.IsNotExist(err) {
		log.Printf("警告: DKIM私钥文件不存在: %s", config.DKIM.PrivateKeyPath)
		log.Printf("您可以使用 setup_tunnel.sh 脚本生成DKIM密钥，或者手动创建密钥")
	}
}

// MaskPassword 隐藏密码，只显示前两位和后两位
func MaskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}

package server

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/mhale/smtpd"
	"github.com/nuecms/mailer/config"
	"github.com/nuecms/mailer/mail"
	"github.com/nuecms/mailer/monitoring"
	"github.com/nuecms/mailer/utils"
)

// SetupAndRunSMTPServer 配置并启动SMTP服务器
func SetupAndRunSMTPServer(cfg *config.Config, metrics *monitoring.Metrics, mailQueue chan mail.MailJob) error {
	// 创建认证函数，包含本地连接检查
	authHandler := func(remoteAddr net.Addr, mechanism string, username []byte, password []byte, shared []byte) (bool, error) {
		// 检查连接是否来自本地
		if cfg.Security.AllowLocalOnly && !utils.IsLocalConnection(remoteAddr) {
			log.Printf("拒绝非本地连接: %v", remoteAddr)
			return false, fmt.Errorf("只允许本地连接")
		}

		log.Printf("接收到验证请求，机制: %s, 用户名: %s", mechanism, username)

		// 如果没有配置用户名密码，则接受任何来自本地的认证
		if cfg.DefaultUsername == "" {
			return true, nil
		}

		// 验证用户名密码
		if mechanism == "PLAIN" || mechanism == "LOGIN" {
			return string(username) == cfg.DefaultUsername && string(password) == cfg.DefaultPassword, nil
		} else if mechanism == "CRAM-MD5" {
			// 处理 CRAM-MD5 认证
			expectedDigest := utils.ComputeCRAMMD5(string(shared), cfg.DefaultPassword)
			return string(username) == cfg.DefaultUsername && string(password) == expectedDigest, nil
		}

		// 不支持的验证机制
		log.Printf("不支持的验证机制: %s", mechanism)
		return false, fmt.Errorf("不支持的验证机制: %s", mechanism)
	}

	// 创建邮件处理函数，同样检查本地连接
	mailHandler := func(origin net.Addr, from string, to []string, data []byte) error {
		// 再次检查连接是否来自本地
		if cfg.Security.AllowLocalOnly && !utils.IsLocalConnection(origin) {
			log.Printf("拒绝来自非本地的邮件: %v", origin)
			return fmt.Errorf("只允许本地连接发送邮件")
		}

		// 生成邮件ID
		mailID := utils.GenerateID()

		log.Printf("[%s] 收到邮件: 从 %s 到 %s", mailID, from, utils.SummarizeRecipients(to))

		// 检查公共邮箱发送提示
		for _, recipient := range to {
			if strings.Contains(recipient, "gmail.com") ||
				strings.Contains(recipient, "outlook.com") ||
				strings.Contains(recipient, "hotmail.com") ||
				strings.Contains(recipient, "ethereal.email") {
				log.Printf("[%s] 发送到公共邮箱: %s，如果无法收到邮件，请检查发件人域名是否有SPF记录或转发服务是否有授权",
					mailID, recipient)
			}
		}

		// 检查速率限制
		if cfg.RateLimits.Enabled {
			if !metrics.CheckRateLimit(from, cfg) {
				log.Printf("[%s] 发件人 %s 超过速率限制", mailID, from)
				return fmt.Errorf("发送频率过高，请稍后重试")
			}
		}

		// 将邮件放入队列异步处理
		mailQueue <- mail.MailJob{
			From: from,
			To:   to,
			Data: data,
			ID:   mailID,
		}

		log.Printf("[%s] 邮件已加入队列等待处理", mailID)
		return nil
	}

	// 确保SMTPHost设置为本地地址，如果需要强制本地连接
	if cfg.Security.AllowLocalOnly && cfg.SMTPHost != "127.0.0.1" && cfg.SMTPHost != "localhost" {
		log.Printf("警告: SMTPHost 不是本地地址 (当前值: %s)，已强制改为 127.0.0.1", cfg.SMTPHost)
		cfg.SMTPHost = "127.0.0.1"
	}

	// 启动SMTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	server := smtpd.Server{
		Addr:         addr,
		Handler:      mailHandler,
		Appname:      "Go Mail Server",
		AuthHandler:  authHandler,
		AuthRequired: cfg.Security.RequireAuth || cfg.DefaultUsername != "",
		MaxSize:      10485760, // 限制邮件大小为10MB
		Timeout:      time.Minute * 5, // 设置超时时间为5分钟
	}

	localOnlyMsg := ""
	if cfg.Security.AllowLocalOnly {
		localOnlyMsg = "(仅限本地连接)"
	}

	log.Printf("SMTP服务器启动在 %s %s", addr, localOnlyMsg)

	if cfg.DefaultUsername != "" {
		log.Printf("认证信息：用户名=%s, 密码=%s",
			cfg.DefaultUsername,
			config.MaskPassword(cfg.DefaultPassword))
	}

	return server.ListenAndServe()
}

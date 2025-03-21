package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"time"
	"github.com/nuecms/mailer/config"
	"github.com/nuecms/mailer/utils"
)

// MailJob 表示一个待处理的邮件作业
type MailJob struct {
	From string
	To   []string
	Data []byte
	ID   string
}

// ForwardMail 实现转发邮件到配置的SMTP服务器
func ForwardMail(cfg *config.Config, from string, to []string, data []byte) error {
	// 添加批处理功能，每批最多发送20个收件人
	batchSize := 20
	if cfg != nil && cfg.BatchSize > 0 {
		batchSize = cfg.BatchSize
	}

	if len(to) > batchSize {
		batches := make([][]string, 0, (len(to)+batchSize-1)/batchSize)
		for i := 0; i < len(to); i += batchSize {
			end := i + batchSize
			if end > len(to) {
				end = len(to)
			}
			batches = append(batches, to[i:end])
		}

		log.Printf("收件人过多，分批发送 (%d 批)", len(batches))

		for i, batch := range batches {
			log.Printf("发送第 %d 批 (共 %d 个收件人)", i+1, len(batch))
			if err := ForwardMailBatch(cfg, from, batch, data); err != nil {
				log.Printf("第 %d 批发送失败: %v", i+1, err)
				return err
			}

			// 批次间延迟
			if cfg != nil && i < len(batches)-1 && cfg.BatchDelay > 0 {
				time.Sleep(time.Duration(cfg.BatchDelay) * time.Millisecond)
			}
		}

		return nil
	}

	return ForwardMailBatch(cfg, from, to, data)
}

// ForwardMailBatch 实际执行邮件转发功能
func ForwardMailBatch(cfg *config.Config, from string, to []string, data []byte) error {
	// 如果没有配置，使用默认配置
	var forwardHost, forwardUsername, forwardPassword string
	var forwardPort int
	var forwardSSL bool

	if cfg != nil {
		forwardHost = cfg.ForwardHost
		forwardPort = cfg.ForwardPort
		forwardUsername = cfg.ForwardUsername
		forwardPassword = cfg.ForwardPassword
		forwardSSL = cfg.ForwardSSL
	} else {
		// 加载默认配置
		defaultConfig, err := config.Load("config.json")
		if err != nil {
			return fmt.Errorf("无法加载默认配置: %v", err)
		}
		forwardHost = defaultConfig.ForwardHost
		forwardPort = defaultConfig.ForwardPort
		forwardUsername = defaultConfig.ForwardUsername
		forwardPassword = defaultConfig.ForwardPassword
		forwardSSL = defaultConfig.ForwardSSL
	}

	// 准备SMTP地址
	addr := fmt.Sprintf("%s:%d", forwardHost, forwardPort)
	log.Printf("连接到SMTP服务器: %s", addr)

	// 显示详细的调试日志
	if forwardUsername != "" {
		log.Printf("使用认证信息: 用户名=%s", forwardUsername)
	} else {
		log.Printf("未配置转发认证信息")
	}

	// 显示更多邮件信息
	dataLen := len(data)
	previewLen := utils.Min(200, dataLen)
	log.Printf("邮件头部预览: %s", string(data[:previewLen]))

	// 增加指数退避重试机制
	retryCount := 3
	backoff := time.Second

	// 重构代码以避免使用goto跳过变量声明
	for i := 0; i < retryCount; i++ {
		err := tryToSendMail(forwardHost, forwardPort, forwardUsername, forwardPassword, forwardSSL, addr, from, to, data)
		if err == nil {
			// 成功发送
			log.Printf("成功转发邮件给 %v", utils.SummarizeRecipients(to))
			return nil
		}

		// 连接错误可能是暂时性的，尝试重试
		if i < retryCount-1 {
			log.Printf("尝试发送失败 (%d/%d)，将在 %v 后重试: %v", 
				i+1, retryCount, backoff, err)
			time.Sleep(backoff)
			backoff *= 2 // 指数递增
		} else {
			// 最后一次尝试也失败
			return fmt.Errorf("多次尝试后发送失败: %v", err)
		}
	}

	// 不应该到达这里，但为了编译器不报错
	return fmt.Errorf("发送失败")
}

// tryToSendMail 尝试发送一封邮件，封装了单次发送的逻辑
func tryToSendMail(forwardHost string, forwardPort int, forwardUsername, forwardPassword string, 
	forwardSSL bool, addr string, from string, to []string, data []byte) error {
	
	// 创建SMTP客户端连接
	var client *smtp.Client
	var err error

	if forwardSSL {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			ServerName: forwardHost,
			// 在生产环境中应该设置为true
			InsecureSkipVerify: false,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("无法创建TLS连接: %v", err)
		}

		client, err = smtp.NewClient(conn, forwardHost)
		if err != nil {
			return fmt.Errorf("无法创建SMTP客户端: %v", err)
		}
	} else {
		// 使用普通连接
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("无法连接到SMTP服务器: %v", err)
		}

		// 如果服务器支持，启用TLS
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: forwardHost}
			if err = client.StartTLS(tlsConfig); err != nil {
				log.Printf("启用TLS失败: %v", err)
				// 继续，不要返回错误
			}
		}
	}
	defer client.Close()

	// 认证
	if forwardUsername != "" && forwardPassword != "" {
		auth := smtp.PlainAuth("", forwardUsername, forwardPassword, forwardHost)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %v", err)
		}
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	recipientFailCount := 0
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			log.Printf("设置收件人 %s 失败: %v", recipient, err)
			recipientFailCount++
			// 继续其他收件人，不要立即返回错误
		}
	}

	// 如果所有收件人都失败，则视为整体失败
	if recipientFailCount == len(to) {
		return fmt.Errorf("所有收件人设置失败")
	}

	// 发送数据
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送数据失败: %v", err)
	}

	if _, err = w.Write(data); err != nil {
		return fmt.Errorf("写入邮件数据失败: %v", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("完成数据发送失败: %v", err)
	}

	// 结束会话
	err = client.Quit()
	if err != nil {
		log.Printf("关闭SMTP连接失败: %v", err)
		// 不要返回错误，因为邮件已经发送
	}

	return nil
}

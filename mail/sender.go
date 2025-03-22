package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"time"
	"github.com/nuecms/mailer/config"
	"github.com/nuecms/mailer/utils"
	"sort"
)

// MailJob 表示一个待处理的邮件作业
type MailJob struct {
	From string
	To   []string
	Data []byte
	ID   string
}

// ProcessMail 处理邮件发送，按优先级尝试不同方式
// 1. 直接外发(如果配置了直接外发且配置有效)
// 2. SMTP转发(如果配置了SMTP转发且配置有效)
// 3. 本地存储(作为最后的保底方案)
func ProcessMail(cfg *config.Config, from string, to []string, data []byte) error {
	// 如果启用了DKIM，对邮件进行签名
	if cfg.DKIM != nil && cfg.DKIM.Enabled {
		signedData, err := SignWithDKIM(cfg, data)
		if err != nil {
			log.Printf("DKIM签名失败: %v, 将使用未签名邮件继续", err)
		} else {
			data = signedData
			log.Printf("邮件已成功添加DKIM签名")
		}
	}

	// 尝试直接外发
	if cfg.DirectDelivery != nil && cfg.DirectDelivery.Enabled {
		log.Printf("尝试直接发送邮件到目标服务器")
		err := SendMailDirect(cfg, from, to, data)
		if err == nil {
			log.Printf("直接发送邮件成功")
			return nil
		}
		log.Printf("直接发送邮件失败: %v, 将尝试SMTP转发", err)
	}

	// 尝试SMTP转发
	// 检查是否有配置转发提供商或传统的转发设置
	hasForwardingConfig := (cfg.ForwardSMTP && len(cfg.ForwardProviders) > 0) || 
						 (cfg.ForwardSMTP && cfg.ForwardHost != "")
	
	if hasForwardingConfig {
		log.Printf("尝试通过SMTP转发邮件")
		err := ForwardMail(cfg, from, to, data)
		if err == nil {
			log.Printf("SMTP转发邮件成功")
			return nil
		}
		log.Printf("SMTP转发邮件失败: %v, 将保存到本地", err)
	}

	// 最后保存到本地
	log.Printf("保存邮件到本地文件系统")
	return SaveMailLocally(from, to, data)
}

// SendMailDirect 尝试直接将邮件发送到目标邮件服务器
func SendMailDirect(cfg *config.Config, from string, to []string, data []byte) error {
	if cfg.DirectDelivery == nil || !cfg.DirectDelivery.Enabled {
		return fmt.Errorf("直接发送功能未启用")
	}

	log.Printf("正在尝试直接发送邮件到收件人服务器")

	// 按域名分组收件人
	domainRecipients := make(map[string][]string)
	for _, recipient := range to {
		domain := utils.ExtractDomain(recipient)
		if domain == "" {
			log.Printf("无法从 %s 提取域名，跳过", recipient)
			continue
		}
		domainRecipients[domain] = append(domainRecipients[domain], recipient)
	}

	// 为每个域名解析MX记录并发送
	successCount := 0
	for domain, recipients := range domainRecipients {
		mxRecords, err := utils.LookupMX(domain)
		if err != nil || len(mxRecords) == 0 {
			log.Printf("无法解析域名 %s 的MX记录: %v, 跳过", domain, err)
			continue
		}

		// 尝试连接到每个MX服务器，直到成功
		delivered := false
		for _, mx := range mxRecords {
			host := mx.Host
			// 确保主机名没有尾随的点
			if host[len(host)-1] == '.' {
				host = host[:len(host)-1]
			}
			port := 25 // 标准SMTP端口

			addr := fmt.Sprintf("%s:%d", host, port)
			log.Printf("尝试连接到MX服务器: %s 发送给 %v", addr, utils.SummarizeRecipients(recipients))

			// 尝试发送
			err := trySendMailToServer(cfg, from, recipients, data, host, port)
			if err == nil {
				log.Printf("成功直接发送邮件到 %s 的MX服务器", domain)
				successCount += len(recipients)
				delivered = true
				break
			}
			log.Printf("发送到 %s 的MX服务器失败: %v, 尝试下一个", host, err)
		}

		if !delivered {
			log.Printf("无法发送到 %s 的任何MX服务器", domain)
		}
	}

	// 如果部分成功，部分失败，视为整体成功
	if successCount > 0 {
		if successCount < len(to) {
			log.Printf("部分直接发送成功: %d/%d 收件人", successCount, len(to))
		}
		return nil
	}

	return fmt.Errorf("所有直接发送尝试均失败")
}

// trySendMailToServer 尝试将邮件直接发送到指定的邮件服务器
func trySendMailToServer(cfg *config.Config, from string, to []string, data []byte, host string, port int) error {
	// 创建 SMTP 客户端连接
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := smtp.Dial(addr)
	if (err != nil) {
		return fmt.Errorf("无法连接到服务器: %v", err)
	}
	defer client.Close()

	// 如果配置了EHLO域名，使用它；否则使用发件人域名
	ehlo := utils.ExtractDomain(from)
	if cfg.DirectDelivery.EhloDomain != "" {
		ehlo = cfg.DirectDelivery.EhloDomain
	}

	if err := client.Hello(ehlo); err != nil {
		return fmt.Errorf("EHLO 失败: %v", err)
	}

	// 如果服务器支持，尝试启用TLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: cfg.DirectDelivery.InsecureSkipVerify,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			log.Printf("启用 TLS 失败: %v, 继续不使用TLS", err)
		}
	}

	// 设置发件人
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	recipientFailCount := 0
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			log.Printf("设置收件人 %s 失败: %v", recipient, err)
			recipientFailCount++
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

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("写入邮件数据失败: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("完成数据发送失败: %v", err)
	}

	// 关闭连接
	if err := client.Quit(); err != nil {
		log.Printf("关闭连接失败: %v", err)
		// 不返回错误，因为邮件已经发送
	}

	return nil
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
	// 首先确保配置被转换为多提供商格式
	if cfg != nil {
		config.ConvertLegacyConfig(cfg)
	}
	
	// 获取提供商列表
	var providers []config.SMTPProvider
	if cfg != nil && cfg.ForwardSMTP && len(cfg.ForwardProviders) > 0 {
		providers = cfg.ForwardProviders
	} else if cfg != nil && cfg.ForwardSMTP && cfg.ForwardHost != "" {
		// 旧式配置兼容 (冗余检查)
		providers = []config.SMTPProvider{
			{
				Host:     cfg.ForwardHost,
				Port:     cfg.ForwardPort,
				Username: cfg.ForwardUsername,
				Password: cfg.ForwardPassword,
				SSL:      cfg.ForwardSSL,
			},
		}
	} else {
		// 加载默认配置
		defaultConfig, err := config.Load("config.json")
		if err != nil {
			return fmt.Errorf("无法加载默认配置: %v", err)
		}
		config.ConvertLegacyConfig(defaultConfig)
		
		// 确保默认配置也检查 forwardSMTP 标志
		if !defaultConfig.ForwardSMTP {
			return fmt.Errorf("SMTP转发功能已禁用，无法转发邮件")
		}
		
		providers = defaultConfig.ForwardProviders
	}
	
	// 如果没有可用的提供商，返回错误
	if len(providers) == 0 {
		return fmt.Errorf("未配置SMTP提供商，无法转发邮件")
	}
	
	// 按照优先级排序提供商
	sort.Slice(providers, func(i, j int) bool {
		// 如果优先级相同，保持原有顺序
		if providers[i].Priority == providers[j].Priority {
			return i < j
		}
		return providers[i].Priority < providers[j].Priority
	})
	
	// 存储错误以便返回最后一个错误
	var lastError error
	
	// 尝试每个提供商
	for i, provider := range providers {
		log.Printf("尝试使用SMTP提供商 #%d: %s", i+1, provider.Host)
		
		// 准备SMTP地址
		addr := fmt.Sprintf("%s:%d", provider.Host, provider.Port)
		log.Printf("连接到SMTP服务器: %s", addr)
		
		// 显示详细的调试日志
		if provider.Username != "" {
			log.Printf("使用认证信息: 用户名=%s", provider.Username)
		} else {
			log.Printf("未配置认证信息")
		}
		
		// 显示更多邮件信息
		dataLen := len(data)
		previewLen := utils.Min(200, dataLen)
		log.Printf("邮件头部预览: %s", string(data[:previewLen]))
		
		// 用当前提供商尝试发送
		err := trySendWithProvider(provider, from, to, data)
		if err == nil {
			// 成功发送
			log.Printf("成功使用提供商 %s 转发邮件给 %v", provider.Host, utils.SummarizeRecipients(to))
			return nil
		}
		
		// 记录错误并尝试下一个提供商
		lastError = fmt.Errorf("提供商 %s 发送失败: %v", provider.Host, err)
		log.Printf("使用提供商 %s 发送失败: %v, 尝试下一个提供商", provider.Host, err)
	}
	
	// 所有提供商都失败
	return fmt.Errorf("所有SMTP提供商均发送失败，最后错误: %v", lastError)
}

// trySendWithProvider 使用指定的SMTP提供商尝试发送邮件
func trySendWithProvider(provider config.SMTPProvider, from string, to []string, data []byte) error {
	// 增加指数退避重试机制
	retryCount := 3
	backoff := time.Second
	
	// 重试循环
	for i := 0; i < retryCount; i++ {
		err := tryToSendMailWithProvider(provider, from, to, data)
		if err == nil {
			// 成功发送
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

// tryToSendMailWithProvider 基于提供商配置尝试发送邮件
func tryToSendMailWithProvider(provider config.SMTPProvider, from string, to []string, data []byte) error {
	// 创建SMTP客户端连接
	var client *smtp.Client
	var err error
	
	// 准备SMTP地址
	addr := fmt.Sprintf("%s:%d", provider.Host, provider.Port)
	
	if provider.SSL {
		// 使用TLS连接
		tlsConfig := &tls.Config{
			ServerName: provider.Host,
			// 在生产环境中应该设置为true
			InsecureSkipVerify: false,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("无法创建TLS连接: %v", err)
		}

		client, err = smtp.NewClient(conn, provider.Host)
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
			tlsConfig := &tls.Config{ServerName: provider.Host}
			if err = client.StartTLS(tlsConfig); err != nil {
				log.Printf("启用TLS失败: %v", err)
				// 继续，不要返回错误
			}
		}
	}
	defer client.Close()

	// 认证
	if provider.Username != "" && provider.Password != "" {
		auth := smtp.PlainAuth("", provider.Username, provider.Password, provider.Host)
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

// SignWithDKIM 使用DKIM对邮件进行签名
func SignWithDKIM(cfg *config.Config, data []byte) ([]byte, error) {
	if cfg.DKIM == nil || !cfg.DKIM.Enabled {
		return nil, fmt.Errorf("DKIM未启用")
	}

	// 创建DKIM签名器
	signer, err := NewDKIMSigner(DKIMOptions{
		Domain:            cfg.DKIM.Domain,
		Selector:          cfg.DKIM.Selector,
		PrivateKeyPath:    cfg.DKIM.PrivateKeyPath,
		HeadersToSign:     cfg.DKIM.HeadersToSign,
		SignatureExpireIn: cfg.DKIM.SignatureExpiry,
	})
	if err != nil {
		return nil, fmt.Errorf("创建DKIM签名器失败: %v", err)
	}

	// 对邮件进行签名
	return signer.SignMessage(data)
}

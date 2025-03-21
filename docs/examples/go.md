# Go 使用示例

本文档提供了使用 Go 语言连接 Go Mail Server 发送邮件的示例代码。

## 使用标准库发送邮件

Go 标准库包含了 `net/smtp` 包，可以用于发送邮件。

### 基本邮件发送

```go
package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func main() {
	// 配置 SMTP 参数
	smtpServer := "localhost"
	smtpPort := 25
	smtpUser := "noreply@example.com"
	smtpPassword := "your-password"
	
	// 邮件参数
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	subject := "Go 测试邮件"
	body := "这是使用 Go 语言发送的测试邮件。"
	
	// 构建邮件内容
	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, strings.Join(to, ", "), subject, body))
	
	// 认证
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)
	
	// 发送邮件
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", smtpServer, smtpPort),
		auth,
		from,
		to,
		message,
	)
	
	if err != nil {
		log.Fatalf("发送邮件失败: %v", err)
	}
	
	log.Println("邮件发送成功")
}
```

### 发送 HTML 邮件

```go
package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func main() {
	// 配置 SMTP 参数
	smtpServer := "localhost"
	smtpPort := 25
	smtpUser := "noreply@example.com"
	smtpPassword := "your-password"
	
	// 邮件参数
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	subject := "HTML 测试邮件"
	
	// HTML 内容
	htmlBody := `
	<html>
		<body>
			<h1>Go Mail Server 测试</h1>
			<p>这是一封 <b>HTML</b> 格式的测试邮件。</p>
			<p>您可以包含链接: <a href="https://example.com">Example</a></p>
		</body>
	</html>
	`
	
	// 构建邮件头部
	headers := map[string]string{
		"From":         from,
		"To":           strings.Join(to, ", "),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}
	
	// 构建完整邮件
	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody
	
	// 认证
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)
	
	// 发送邮件
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", smtpServer, smtpPort),
		auth,
		from,
		to,
		[]byte(message),
	)
	
	if err != nil {
		log.Fatalf("发送邮件失败: %v", err)
	}
	
	log.Println("HTML 邮件发送成功")
}
```

## 使用第三方库发送邮件

Go 社区中有多个邮件发送库，这里以 [gomail](https://github.com/go-gomail/gomail) 为例。

### 安装 gomail

```bash
go get gopkg.in/gomail.v2
```

### 使用 gomail 发送带附件的邮件

```go
package main

import (
	"fmt"
	"log"
	
	"gopkg.in/gomail.v2"
)

func main() {
	// 创建新邮件
	m := gomail.NewMessage()
	m.SetHeader("From", "sender@example.com")
	m.SetHeader("To", "recipient1@example.com", "recipient2@example.com")
	m.SetHeader("Subject", "带附件的邮件")
	
	// 设置 HTML 正文
	m.SetBody("text/html", "<p>这是邮件正文</p><p>请查看附件。</p>")
	
	// 添加附件
	m.Attach("./document.pdf")
	
	// 创建发送器
	d := gomail.NewDialer("localhost", 25, "noreply@example.com", "your-password")
	
	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("发送邮件失败: %v", err)
	}
	
	log.Println("邮件已成功发送")
}
```

### 批量发送邮件

```go
package main

import (
	"fmt"
	"log"
	"sync"
	
	"gopkg.in/gomail.v2"
)

func main() {
	// 收件人列表
	recipients := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
		"user4@example.com",
		"user5@example.com",
	}
	
	// 创建发送器
	d := gomail.NewDialer("localhost", 25, "noreply@example.com", "your-password")
	
	// 建立连接
	s, err := d.Dial()
	if err != nil {
		log.Fatalf("无法连接到SMTP服务器: %v", err)
	}
	defer s.Close()
	
	// 使用 WaitGroup 等待所有邮件发送完成
	var wg sync.WaitGroup
	
	for i, recipient := range recipients {
		wg.Add(1)
		
		go func(i int, recipient string) {
			defer wg.Done()
			
			m := gomail.NewMessage()
			m.SetHeader("From", "sender@example.com")
			m.SetHeader("To", recipient)
			m.SetHeader("Subject", fmt.Sprintf("批量测试邮件 #%d", i+1))
			m.SetBody("text/plain", fmt.Sprintf("这是发送给 %s 的测试邮件 #%d", recipient, i+1))
			
			if err := gomail.Send(s, m); err != nil {
				log.Printf("发送到 %s 失败: %v", recipient, err)
			} else {
				log.Printf("成功发送到: %s", recipient)
			}
		}(i, recipient)
	}
	
	// 等待所有邮件发送完成
	wg.Wait()
	log.Println("批量邮件发送完成")
}
```

## 使用 Cloudflare Tunnel 访问

如果您使用 Cloudflare Tunnel 配置了外部访问：

```go
package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func main() {
	// 使用 Tunnel 域名
	smtpServer := "smtp.yourdomain.com" // 替换为您的 Tunnel 域名
	smtpPort := 25
	smtpUser := "noreply@example.com"
	smtpPassword := "your-password"
	
	// 邮件参数
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	subject := "通过 Tunnel 发送的邮件"
	body := "这封邮件通过 Cloudflare Tunnel 发送。"
	
	// 构建邮件内容
	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, strings.Join(to, ", "), subject, body))
	
	// 认证
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)
	
	// 发送邮件
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", smtpServer, smtpPort),
		auth,
		from,
		to,
		message,
	)
	
	if err != nil {
		log.Fatalf("发送邮件失败: %v", err)
	}
	
	log.Println("通过 Tunnel 发送邮件成功")
}
```

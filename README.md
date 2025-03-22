# Go Mail Server

Go Mail Server 是一个轻量级 SMTP 服务器，专为内部应用程序提供邮件发送功能。它可以作为本地 SMTP 代理，接收应用程序的邮件请求，然后将其转发到外部 SMTP 服务器或保存到本地。

[![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 功能特点

- **仅本地连接**：默认只接受来自本地的连接，增强安全性
- **直接发送**：支持直接发送邮件到收件人邮件服务器，无需第三方SMTP
- **邮件转发**：将邮件转发至外部 SMTP 服务器（如 Gmail、阿里云等）
- **多提供商支持**：配置多个SMTP提供商实现自动故障转移
- **批量处理**：支持批量邮件发送，自动分批处理大量收件人
- **异步处理**：后台处理邮件发送，提高响应速度
- **故障恢复**：自动保存失败邮件，定期重试
- **监控支持**：提供健康检查和指标接口
- **速率限制**：控制每个发件人的发送频率
- **DKIM 支持**：可配置 DKIM 签名提高送达率

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/nuecms/mailer.git
cd mailer

# 创建配置文件
cp config.example.json config.json

# 编辑配置文件
# 根据需要修改 config.json

# 构建项目
go build -o mailer

# 运行服务
./mailer -config config.json
```

### Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d

# 或者手动运行容器
docker run -d --name mailer \
  -p 127.0.0.1:25:25 \
  -p 127.0.0.1:8025:8025 \
  -v $(pwd)/config.json:/app/config.json \
  -v $(pwd)/emails:/app/emails \
  nuecms/mailer:latest
```

## 配置选项

基本配置示例：

```json
{
  "smtpHost": "127.0.0.1",
  "smtpPort": 25,
  "defaultUsername": "noreply@example.com",
  "defaultPassword": "your-strong-password",
  
  "directDelivery": {
    "enabled": true,
    "ehloDomain": "example.com",
    "insecureSkipVerify": false,
    "retryCount": 3
  },
  
  "forwardSMTP": true,
  "forwardHost": "smtp.gmail.com",
  "forwardPort": 587,
  "forwardUsername": "your-email@gmail.com",
  "forwardPassword": "your-app-password",
  "forwardSSL": false,
  
  "dkim": {
    "enabled": true,
    "domain": "example.com",
    "selector": "mail",
    "privateKeyPath": "keys/example.com/mail.private"
  },
  
  "batchSize": 20,
  "batchDelay": 1000,
  "enableHealthCheck": true,
  "healthCheckPort": 8025,
  
  "rateLimits": {
    "enabled": true,
    "maxPerHour": 500,
    "maxPerDay": 2000
  },
  
  "security": {
    "allowLocalOnly": true,
    "logAllEmails": true,
    "requireAuth": true
  }
}
```

## 使用方法

### 从应用程序发送邮件

您可以使用任何支持 SMTP 的库或框架连接到 Go Mail Server：

```python
# Python 示例
import smtplib
from email.message import EmailMessage

msg = EmailMessage()
msg.set_content('邮件内容')
msg['Subject'] = '测试邮件'
msg['From'] = 'sender@example.com'
msg['To'] = 'recipient@example.com'

# 连接到本地 SMTP 服务器
s = smtplib.SMTP('localhost', 25)
s.login('noreply@example.com', 'your-strong-password') # 配置文件中的凭据
s.send_message(msg)
s.quit()
```

```javascript
// Node.js 示例 (使用 nodemailer)
const nodemailer = require('nodemailer');

let transporter = nodemailer.createTransport({
  host: 'localhost',
  port: 25,
  auth: {
    user: 'noreply@example.com',  // 配置文件中的凭据
    pass: 'your-strong-password'
  }
});

let mailOptions = {
  from: 'sender@example.com',
  to: 'recipient@example.com',
  subject: '测试邮件',
  text: '邮件内容'
};

transporter.sendMail(mailOptions);
```

### 查看状态和指标

服务启动后，可以通过 HTTP 访问健康检查和指标：

```bash
# 查看健康状态
curl http://localhost:8025/health

# 查看性能指标
curl http://localhost:8025/metrics

# 手动触发失败邮件重试
curl -X POST http://localhost:8025/admin/retry-failed
```

## 文档

完整文档请访问我们的[在线文档站点](https://mailer.nuecms.com/)，其中包含：

- [部署指南](https://mailer.nuecms.com/guides/deployment.html)
- [直接发送模式](https://mailer.nuecms.com/guides/direct-delivery.html)
- [配置详解](https://mailer.nuecms.com/guides/configuration.html)
- [高级功能](https://mailer.nuecms.com/guides/advanced-features.html)
- [DKIM 配置](https://mailer.nuecms.com/guides/dkim-setup.html)
- [性能优化](https://mailer.nuecms.com/guides/optimization.html)
- [故障排查](https://mailer.nuecms.com/guides/troubleshooting.html)

## 监控与维护

### 日志文件

日志默认输出到标准输出。在生产环境中，建议配置日志重定向：

```bash
./mailer -config config.json > mailer.log 2>&1
```

### 健康检查

健康检查端点返回服务状态和详细指标：

```json
{
  "status": "ok",
  "timestamp": "2023-04-01T12:34:56Z",
  "details": {
    "disk_space_available_mb": 8192,
    "queued_emails": 5,
    "failed_emails": 2,
    "total_emails_processed": 1250,
    "success_rate": 99.8
  }
}
```

## 贡献

欢迎提交 Pull Request 或 Issue！在提交代码前，请确保：

1. 代码遵循 Go 标准格式（使用 `go fmt`）
2. 添加必要的测试
3. 更新相关文档

## 许可证

本项目采用 MIT 许可证 - 详情请见 [LICENSE](LICENSE) 文件。

## 开发者

- [NueCMS 团队](https://github.com/nuecms)

## 致谢

- [SMTPD](https://github.com/mhale/smtpd) - 提供 SMTP 服务器基础功能

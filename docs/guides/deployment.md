# Go Mail Server 部署指南

本文档提供了 Go Mail Server 在各种环境中的完整部署指南，涵盖从开发环境到生产环境的多种部署场景。

## 目录

- [Go Mail Server 部署指南](#go-mail-server-部署指南)
  - [目录](#目录)
  - [简介](#简介)
  - [部署前准备](#部署前准备)
    - [系统要求](#系统要求)
    - [前置检查](#前置检查)
  - [基本部署方式](#基本部署方式)
    - [本地开发环境](#本地开发环境)
    - [服务器直接部署](#服务器直接部署)
    - [Docker 容器部署](#docker-容器部署)
    - [Cloudflare Tunnel 部署](#cloudflare-tunnel-部署)
  - [配置说明](#配置说明)
  - [高可用部署](#高可用部署)
    - [多实例部署](#多实例部署)
    - [多提供商故障转移](#多提供商故障转移)
  - [安全加固](#安全加固)
    - [基本安全措施](#基本安全措施)
    - [高级安全措施](#高级安全措施)
  - [监控方案](#监控方案)
    - [内置监控](#内置监控)
    - [集成第三方监控](#集成第三方监控)
  - [故障排查](#故障排查)
    - [常见问题](#常见问题)
    - [日志解读](#日志解读)
  - [生产环境最佳实践](#生产环境最佳实践)

## 简介

Go Mail Server 是一个轻量级的邮件发送服务，专为内部应用程序提供 SMTP 服务。它的主要特点包括：

- **仅本地连接**：默认只接受来自本地的连接，增强安全性
- **邮件转发**：可将邮件转发至外部 SMTP 服务器（如 Gmail、阿里云等）
- **批量处理**：支持批量邮件发送，自动分批处理大量收件人
- **异步处理**：后台处理邮件发送，提高响应速度
- **故障恢复**：自动保存失败邮件，定期重试
- **监控支持**：提供健康检查和指标接口

## 部署前准备

### 系统要求

- Go 1.16+ (仅构建时需要)
- 任何支持 Docker 的操作系统（如果使用 Docker 部署）
- 至少 512MB RAM
- 至少 1GB 存储空间（用于邮件存储和日志）

### 前置检查

1. **网络连接**：确保服务器可以访问外部 SMTP 服务（如果使用转发模式）
2. **端口检查**：确保端口未被占用（默认使用 25 端口和 8025 端口）
   ```bash
   # 检查端口是否被占用
   sudo netstat -tuln | grep -E ':(25|8025)'
   ```
3. **防火墙配置**：如果需要远程访问，确保防火墙允许相关端口

## 基本部署方式

### 本地开发环境

最简单的部署方式，适合开发和测试：

```bash
# 克隆仓库
git clone https://github.com/nuecms/mailer.git
cd mailer

# 复制配置文件
cp config.example.json config.json

# 编辑配置
nano config.json

# 构建应用
go build -o mailserver

# 运行服务
./mailserver -config config.json
```

### 服务器直接部署

适合小型生产环境：

1. **准备服务器**
   ```bash
   # 创建应用目录
   mkdir -p /opt/mailer
   cd /opt/mailer
   
   # 创建用户（可选但推荐）
   sudo useradd -r -s /bin/false mailer
   ```

2. **复制文件**
   将编译好的二进制文件和配置文件复制到服务器

3. **创建 systemd 服务**
   ```
   # /etc/systemd/system/mailer.service
   [Unit]
   Description=Go Mail Server
   After=network.target
   
   [Service]
   Type=simple
   User=mailer
   Group=mailer
   WorkingDirectory=/opt/mailer
   ExecStart=/opt/mailer/mailserver -config /opt/mailer/config.json
   Restart=on-failure
   RestartSec=10
   
   [Install]
   WantedBy=multi-user.target
   ```

4. **启动服务**
   ```bash
   sudo systemctl enable mailer
   sudo systemctl start mailer
   ```

5. **检查状态**
   ```bash
   sudo systemctl status mailer
   journalctl -u mailer -f
   ```

### Docker 容器部署

适合现代化部署环境：

1. **准备 Docker 环境**
   确保已安装 Docker 和 Docker Compose

2. **使用预构建镜像**
   ```bash
   # 创建项目目录
   mkdir -p ~/mailer-docker
   cd ~/mailer-docker
   
   # 创建配置目录
   mkdir -p config emails
   
   # 创建配置文件
   cp /path/to/config.json ./config/
   
   # 创建 docker-compose.yml
   cat > docker-compose.yml << 'EOF'
   version: '3'
   
   services:
     mailer:
       image: nuecms/mailer:latest
       restart: always
       ports:
         - "127.0.0.1:25:25"      # SMTP 端口（仅本地访问）
         - "127.0.0.1:8025:8025"  # 健康检查 HTTP 服务端口
       volumes:
         - ./config/config.json:/app/config.json
         - ./emails:/app/emails
       environment:
         - TZ=Asia/Shanghai
       healthcheck:
         test: ["CMD", "curl", "-f", "http://localhost:8025/health"]
         interval: 30s
         timeout: 5s
         retries: 3
   EOF
   
   # 启动服务
   docker-compose up -d
   ```

3. **或者自行构建镜像**
   ```bash
   # 克隆代码
   git clone https://github.com/nuecms/mailer.git
   cd mailer
   
   # 构建镜像
   docker build -t mailer:latest .
   
   # 启动容器
   docker run -d --name mailer \
     -p 127.0.0.1:25:25 \
     -p 127.0.0.1:8025:8025 \
     -v $(pwd)/config.json:/app/config.json \
     -v $(pwd)/emails:/app/emails \
     mailer:latest
   ```

### Cloudflare Tunnel 部署

适合需要从外部访问但没有公网 IP 的环境：

1. **设置 Cloudflare Tunnel**

   参考我们的 [Cloudflare Tunnel 设置指南](/guides/cloudflare-tunnel) 进行详细配置。

2. **配置邮件服务**

   ```json
   {
     "smtpHost": "127.0.0.1",
     "smtpPort": 25,
     "security": {
       "allowLocalOnly": false,  // 允许来自 Cloudflare Tunnel 的连接
       "requireAuth": true       // 但仍然需要认证
     },
     // ...其他标准配置
   }
   ```

3. **在应用中使用**

   ```
   SMTP 服务器: smtp.yourdomain.com (你的 Cloudflare Tunnel 域名)
   端口: 25
   用户名: 你的配置用户名
   密码: 你的配置密码
   ```

## 配置说明

主要配置项说明：

```json
{
  "smtpHost": "127.0.0.1",        // SMTP 服务器监听地址，建议保持本地
  "smtpPort": 25,                 // SMTP 服务器监听端口
  "defaultUsername": "user",      // SMTP 认证用户名
  "defaultPassword": "password",  // SMTP 认证密码
  
  "forwardSMTP": true,            // 是否启用转发
  "forwardHost": "smtp.gmail.com", // 转发 SMTP 服务器地址
  "forwardPort": 587,             // 转发 SMTP 服务器端口
  "forwardUsername": "your@gmail.com", // 转发 SMTP 用户名
  "forwardPassword": "app-password",   // 转发 SMTP 密码
  "forwardSSL": false,            // 是否使用 SSL 连接转发服务器
  
  "batchSize": 20,                // 每批发送的最大收件人数
  "batchDelay": 1000,             // 批次间延迟（毫秒）
  "enableHealthCheck": true,      // 是否启用健康检查
  "healthCheckPort": 8025,        // 健康检查端口
  
  "rateLimits": {                 // 速率限制设置
    "enabled": true,
    "maxPerHour": 500,            // 每小时每发件人限制
    "maxPerDay": 2000             // 每天每发件人限制
  },
  
  "security": {                   // 安全设置
    "allowLocalOnly": true,       // 是否只允许本地连接
    "logAllEmails": true,         // 是否记录所有邮件
    "requireAuth": true           // 是否要求认证
  }
}
```

## 高可用部署

对于需要高可用性的环境，可以采用以下部署方式：

### 多实例部署

1. **在多台服务器上部署**

   部署 2-3 台服务器，每台都运行 Go Mail Server 实例

2. **使用负载均衡**

   配置 Nginx 或 HAProxy 作为负载均衡器，分发流量到多个实例

3. **示例 HAProxy 配置**

   ```
   frontend mail_frontend
       bind 127.0.0.1:25
       mode tcp
       default_backend mail_servers
   
   backend mail_servers
       mode tcp
       balance roundrobin
       server mail1 10.0.0.1:25 check
       server mail2 10.0.0.2:25 check
       server mail3 10.0.0.3:25 check
   ```

### 多提供商故障转移

配置多个 SMTP 服务提供商，当主要提供商不可用时自动切换：

```json
{
  "forwardSMTP": true,
  "forwardProviders": [
    {
      "host": "smtp.primary.com",
      "port": 587,
      "username": "user@primary.com",
      "password": "password1",
      "ssl": false
    },
    {
      "host": "smtp.backup.com",
      "port": 587,
      "username": "user@backup.com",
      "password": "password2",
      "ssl": false
    }
  ]
}
```

*注意：多提供商功能尚未实现，这是规划中的功能。*

## 安全加固

### 基本安全措施

1. **强密码**
   设置复杂的认证密码，定期更换

2. **仅本地连接**
   保持 `allowLocalOnly: true`，除非确实需要外部访问

3. **防火墙规则**
   ```bash
   # 仅允许本地访问 SMTP 端口
   sudo ufw allow from 127.0.0.1 to any port 25
   ```

### 高级安全措施

1. **容器沙盒化**
   使用 Docker 部署时，限制容器权限

2. **日志审计**
   定期审查日志文件，检查异常活动

3. **密钥管理**
   使用环境变量或外部密钥管理系统存储敏感信息，而不是明文配置文件

## 监控方案

### 内置监控

访问健康检查接口：

```bash
# 健康状态
curl http://localhost:8025/health

# 指标数据
curl http://localhost:8025/metrics
```

### 集成第三方监控

1. **Prometheus + Grafana**

   创建 Prometheus 抓取配置：
   
   ```yaml
   scrape_configs:
     - job_name: 'mailer'
       metrics_path: '/metrics'
       static_configs:
         - targets: ['localhost:8025']
   ```

2. **简单监控脚本**

   ```bash
   #!/bin/bash
   # 简单健康检查
   response=$(curl -s http://localhost:8025/health)
   status=$(echo $response | jq -r '.status')
   
   if [ "$status" != "ok" ]; then
     echo "邮件服务异常: $response"
     # 发送告警
   fi
   
   # 检查队列积压
   failed=$(echo $response | jq -r '.details.failed_emails')
   if [ "$failed" -gt 100]; then
     echo "邮件队列积压: $failed 封邮件等待处理"
     # 发送告警
   fi
   ```

## 故障排查

### 常见问题

1. **无法启动服务**
   
   检查端口占用：
   ```bash
   sudo lsof -i :25
   sudo lsof -i :8025
   ```
   
   检查配置文件格式：
   ```bash
   jq . config.json
   ```

2. **邮件发送失败**
   
   检查转发设置：
   ```bash
   # 测试与转发 SMTP 服务器的连接
   telnet smtp.example.com 587
   
   # 查看邮件日志
   tail -f /opt/mailer/logs/mailer.log
   ```

3. **健康检查服务无法启动**
   
   尝试更改健康检查端口：
   ```json
   {
     "healthCheckPort": 9025  # 尝试不同端口
   }
   ```

### 日志解读

主要日志信息类型：

- `收到邮件`: 成功接收到邮件请求
- `转发邮件到`: 开始尝试转发邮件
- `连接到SMTP服务器`: 正在连接转发服务器
- `成功转发邮件`: 邮件成功发送
- `转发邮件失败`: 无法转发邮件，通常附带具体错误
- `邮件已保存到`: 邮件已保存到本地（正常模式或失败后的备份）
- `尝试发送失败`: 临时错误，将尝试重试

## 生产环境最佳实践

1. **定期备份**
   
   ```bash
   # 备份配置和邮件
   tar -czf mailer-backup-$(date +%Y%m%d).tar.gz /opt/mailer/config.json /opt/mailer/emails
   ```

2. **日志轮转**

   ```
   # /etc/logrotate.d/mailer
   /opt/mailer/logs/*.log {
       daily
       rotate 14
       compress
       delaycompress
       missingok
       notifempty
       create 0640 mailer mailer
   }
   ```

3. **升级策略**
   
   - 先在测试环境验证新版本
   - 备份当前配置和数据
   - 停止服务，更新二进制，启动服务
   - 监控服务状态确保正常

4. **灾难恢复**
   
   保持详细的部署文档，包括：
   - 完整的安装步骤
   - 配置文件位置和说明
   - 数据备份位置
   - 紧急联系人信息

5. **性能调优**
   
   - 合理设置 `batchSize` 和 `batchDelay`
   - 监控磁盘空间，避免空间耗尽
   - 定期处理 `emails/failed` 目录中的失败邮件
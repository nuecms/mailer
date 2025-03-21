---
layout: home
hero:
  name: Go Mail Server
  text: 轻量级 SMTP 邮件发送服务
  tagline: 简单、高效、可靠的邮件发送解决方案
  image:
    src: /images/logo.svg
    alt: Go Mail Server Logo
  actions:
    - theme: brand
      text: 快速开始
      link: /guides/deployment
    - theme: alt
      text: GitHub
      link: https://github.com/nuecms/mailer

features:
  - icon: 🔒
    title: 安全可靠
    details: 默认仅接受本地连接，内置认证机制，支持 DKIM 签名
  - icon: 🚀
    title: 高性能设计
    details: 异步处理，批量发送，自动重试，处理高并发场景
  - icon: 🔍
    title: 监控完善
    details: 内置健康检查和指标接口，轻松集成监控系统
  - icon: 🌐
    title: 外部访问支持
    details: 通过 Cloudflare Tunnel 实现安全的外部访问，无需公网 IP
---

## 什么是 Go Mail Server?

Go Mail Server 是一个轻量级的 SMTP 邮件发送服务，旨在为内部应用程序提供简单、高效、可靠的邮件发送功能。

它可以作为本地 SMTP 代理，接收应用程序的邮件请求，然后将其转发到外部 SMTP 服务器（如 Gmail、阿里云等）或保存到本地文件系统。

## 核心功能

- **本地 SMTP 服务器**：为应用程序提供标准 SMTP 接口
- **邮件转发**：将邮件转发至外部 SMTP 服务器
- **批量处理**：支持批量邮件发送，自动分批处理
- **异步处理**：后台处理邮件发送，提高响应速度
- **故障恢复**：自动保存失败邮件，定期重试
- **监控支持**：提供健康检查和指标接口
- **速率限制**：控制发送频率，防止滥用
- **DKIM 签名**：提高邮件送达率和安全性
- **外部访问**：通过 Cloudflare Tunnel 支持外部访问

## 快速开始

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

## 为什么选择 Go Mail Server?

- **简单易用**：配置简单，部署方便，无需复杂的邮件服务器设置
- **功能完善**：支持批量发送、异步处理、DKIM 签名等高级功能
- **高性能**：基于 Go 语言开发，高并发处理能力强
- **可靠性**：内置故障恢复机制，确保邮件不丢失
- **安全性**：默认仅允许本地连接，内置认证机制
- **可监控**：提供健康检查和指标接口，易于集成监控系统

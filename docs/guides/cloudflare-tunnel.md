# Cloudflare Tunnel 部署指南

本指南将帮助您设置 Cloudflare Tunnel，使您的 Go Mail Server 可以从外部网络安全访问，无需公网 IP 或复杂的网络配置。

## 什么是 Cloudflare Tunnel?

Cloudflare Tunnel 提供了一种安全的方式，将您的本地服务连接到 Cloudflare 网络，使其可以从互联网安全访问，而无需在防火墙上开放端口或配置静态 IP。

对于 Go Mail Server，这意味着您可以：
- 从任何地方安全地发送邮件
- 不需要公网 IP
- 无需配置复杂的网络规则
- 获得 Cloudflare 的 DDoS 保护

## 前提条件

1. 一个 Cloudflare 账户
2. 一个在 Cloudflare 上管理的域名
3. 已安装的 Go Mail Server
4. 安装了 `cloudflared` 客户端工具

## 安装 Cloudflare Tunnel 客户端

### 在 macOS 上安装

```bash
brew install cloudflare/cloudflare/cloudflared
```

### 在 Linux 上安装

```bash
# Debian/Ubuntu
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb

# CentOS/RHEL
curl -L --output cloudflared.rpm https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-x86_64.rpm
sudo rpm -ivh cloudflared.rpm
```

## 使用自动化脚本配置

我们提供了一个自动化脚本来简化配置过程：

```bash
# 下载脚本
curl -O https://raw.githubusercontent.com/nuecms/mailer/main/setup_tunnel.sh
chmod +x setup_tunnel.sh

# 运行脚本
./setup_tunnel.sh
```

脚本将引导您完成整个设置过程，包括：
- 登录到 Cloudflare 账户
- 创建 Tunnel
- 配置 DNS 记录
- 生成 DKIM 密钥
- 启动 Tunnel 服务

## 手动配置步骤

如果您希望手动配置，请按照以下步骤操作：

### 1. 登录到 Cloudflare

```bash
cloudflared login
```

按照提示在浏览器中完成认证流程。

### 2. 创建 Tunnel

```bash
cloudflared tunnel create mailer-tunnel
```

这将创建一个新的 Tunnel 并保存凭证到本地。

### 3. 创建配置文件

创建 `~/.cloudflared/config.yml` 文件：

```yaml
tunnel: <您的Tunnel ID>
credentials-file: /root/.cloudflared/<您的Tunnel ID>.json

ingress:
  - hostname: mail.yourdomain.com
    service: http://localhost:8025
  - hostname: smtp.yourdomain.com
    service: tcp://localhost:25
  - service: http_status:404
```

### 4. 配置 DNS 记录

```bash
cloudflared tunnel route dns mailer-tunnel mail.yourdomain.com
cloudflared tunnel route dns mailer-tunnel smtp.yourdomain.com
```

### 5. 修改 Go Mail Server 配置

编辑 `config.json` 文件，允许外部连接：

```json
{
  "security": {
    "allowLocalOnly": false,
    "requireAuth": true
  }
}
```

### 6. 启动 Tunnel

```bash
cloudflared tunnel run mailer-tunnel
```

## 使用 systemd 设置服务

为了确保 Tunnel 在后台运行并在系统重启后自动启动，可以创建一个 systemd 服务。

创建文件 `/etc/systemd/system/cloudflared-tunnel.service`：

```ini
[Unit]
Description=Cloudflare Tunnel for Go Mail Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/cloudflared tunnel run mailer-tunnel
Restart=on-failure
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable cloudflared-tunnel
sudo systemctl start cloudflared-tunnel
sudo systemctl status cloudflared-tunnel
```

## DNS 记录配置

为了确保邮件服务正常工作，您需要配置几个重要的 DNS 记录：

### MX 记录
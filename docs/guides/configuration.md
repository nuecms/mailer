# 配置详解

Go Mail Server 通过 JSON 配置文件进行配置，本文将详细介绍所有配置选项。

## 基础配置

```json
{
  "smtpHost": "127.0.0.1",
  "smtpPort": 25,
  "defaultUsername": "noreply@example.com",
  "defaultPassword": "your-strong-password"
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `smtpHost` | string | "127.0.0.1" | SMTP 服务器监听地址 |
| `smtpPort` | int | 25 | SMTP 服务器监听端口 |
| `defaultUsername` | string | "" | SMTP 认证用户名 |
| `defaultPassword` | string | "" | SMTP 认证密码 |

## 转发设置

```json
{
  "forwardSMTP": true,
  "forwardHost": "smtp.gmail.com",
  "forwardPort": 587,
  "forwardUsername": "your-email@gmail.com",
  "forwardPassword": "your-app-password",
  "forwardSSL": false
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `forwardSMTP` | bool | false | 是否启用转发 |
| `forwardHost` | string | "" | 转发 SMTP 服务器地址 |
| `forwardPort` | int | 587 | 转发 SMTP 服务器端口 |
| `forwardUsername` | string | "" | 转发 SMTP 认证用户名 |
| `forwardPassword` | string | "" | 转发 SMTP 认证密码 |
| `forwardSSL` | bool | false | 是否使用 SSL 连接转发服务器 |

## 批量发送

```json
{
  "batchSize": 20,
  "batchDelay": 1000
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `batchSize` | int | 20 | 每批最大收件人数量 |
| `batchDelay` | int | 1000 | 批次间延迟（毫秒） |

## 健康检查

```json
{
  "enableHealthCheck": true,
  "healthCheckPort": 8025
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `enableHealthCheck` | bool | true | 是否启用健康检查 |
| `healthCheckPort` | int | 8025 | 健康检查 HTTP 服务端口 |

## 速率限制

```json
{
  "rateLimits": {
    "enabled": true,
    "maxPerHour": 500,
    "maxPerDay": 2000
  }
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `rateLimits.enabled` | bool | false | 是否启用速率限制 |
| `rateLimits.maxPerHour` | int | 500 | 每小时每发件人最大邮件数 |
| `rateLimits.maxPerDay` | int | 2000 | 每天每发件人最大邮件数 |

## 安全设置

```json
{
  "security": {
    "allowLocalOnly": true,
    "logAllEmails": true,
    "requireAuth": true
  }
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `security.allowLocalOnly` | bool | true | 是否只允许本地连接 |
| `security.logAllEmails` | bool | true | 是否记录所有邮件 |
| `security.requireAuth` | bool | true | 是否要求认证 |

## DKIM 配置 (计划中)

```json
{
  "dkim": {
    "enabled": true,
    "domain": "example.com",
    "selector": "mail",
    "privateKeyPath": "/path/to/private.key"
  }
}
```

| 选项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `dkim.enabled` | bool | false | 是否启用 DKIM 签名 |
| `dkim.domain` | string | "" | 签名域名 |
| `dkim.selector` | string | "mail" | DKIM 选择器 |
| `dkim.privateKeyPath` | string | "" | 私钥文件路径 |

## 完整配置示例

```json
{
  "smtpHost": "127.0.0.1",
  "smtpPort": 25,
  "defaultUsername": "noreply@example.com",
  "defaultPassword": "your-strong-password",
  
  "forwardSMTP": true,
  "forwardHost": "smtp.gmail.com",
  "forwardPort": 587,
  "forwardUsername": "your-email@gmail.com",
  "forwardPassword": "your-app-password",
  "forwardSSL": false,
  
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
  },
  
  "dkim": {
    "enabled": false,
    "domain": "example.com",
    "selector": "mail",
    "privateKeyPath": "/path/to/private.key"
  }
}
```

## 环境变量覆盖 (计划中)

所有配置选项都可以通过环境变量覆盖，格式为 `MAILER_` 前缀加上大写配置选项名，例如：

```bash
MAILER_SMTP_HOST=0.0.0.0
MAILER_SMTP_PORT=25
MAILER_DEFAULT_USERNAME=admin
MAILER_DEFAULT_PASSWORD=password
```

嵌套选项使用下划线连接，例如：

```bash
MAILER_RATE_LIMITS_ENABLED=true
MAILER_RATE_LIMITS_MAX_PER_HOUR=500
MAILER_SECURITY_ALLOW_LOCAL_ONLY=true
```

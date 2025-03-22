# 配置详解

Go Mail Server 的配置通过 JSON 文件进行管理，提供了丰富而灵活的配置选项。本指南详细介绍了所有可用的配置项。

## 基本配置

基本配置包括服务器的基本参数和认证信息。

```json
{
  "smtpHost": "127.0.0.1",        // SMTP 服务器监听地址
  "smtpPort": 25,                 // SMTP 服务器监听端口
  "defaultUsername": "noreply@example.com", // SMTP 认证用户名
  "defaultPassword": "your-strong-password" // SMTP 认证密码
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `smtpHost` | 字符串 | SMTP 服务器监听地址，建议保持为本地地址 | `"127.0.0.1"` |
| `smtpPort` | 整数 | SMTP 服务器监听端口 | `25` |
| `defaultUsername` | 字符串 | SMTP 认证用户名 | 无，建议设置 |
| `defaultPassword` | 字符串 | SMTP 认证密码 | 无，建议设置 |

## 转发配置

Go Mail Server 支持将邮件转发到外部 SMTP 服务器。有两种配置方式：多提供商模式（推荐）和传统模式。

### 多提供商模式（推荐）

此模式支持配置多个 SMTP 服务提供商，实现自动故障转移。

```json
{
  "forwardSMTP": true,                 // 总开关，是否启用转发功能
  "forwardProviders": [
    {
      "host": "smtp.primary.com",      // 主要 SMTP 服务器地址
      "port": 587,                     // SMTP 服务器端口
      "username": "user@primary.com",  // 认证用户名
      "password": "password1",         // 认证密码
      "ssl": false,                    // 是否使用 SSL 连接
      "priority": 0                    // 优先级，数字越小优先级越高
    },
    {
      "host": "smtp.backup.com",       // 备用 SMTP 服务器地址
      "port": 587,
      "username": "user@backup.com",
      "password": "password2",
      "ssl": false,
      "priority": 1
    }
  ]
}
```

### 传统模式（向后兼容）

```json
{
  "forwardSMTP": true,                // 是否启用转发
  "forwardHost": "smtp.gmail.com",    // 转发 SMTP 服务器地址
  "forwardPort": 587,                 // 转发 SMTP 服务器端口
  "forwardUsername": "your@gmail.com", // 转发 SMTP 用户名
  "forwardPassword": "app-password",   // 转发 SMTP 密码
  "forwardSSL": false                 // 是否使用 SSL 连接转发服务器
}
```

**注意：** 如果同时存在多提供商配置和传统配置，系统将优先使用多提供商配置。`forwardSMTP` 作为总开关，如果设置为 `false`，则所有转发功能都将被禁用。

## 直接发送配置

直接发送模式允许系统直接将邮件发送到收件人的邮件服务器，无需中间 SMTP 服务器。

```json
{
  "directDelivery": {
    "enabled": true,              // 是否启用直接发送模式
    "ehloDomain": "example.com",  // 用于 EHLO 的域名
    "insecureSkipVerify": false,  // 是否跳过 TLS 验证（不推荐）
    "retryCount": 3               // 发送失败时的重试次数
  }
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `enabled` | 布尔值 | 是否启用直接发送模式 | `false` |
| `ehloDomain` | 字符串 | 用于 EHLO 命令的域名，通常是您的发件域名 | 空，建议设置 |
| `insecureSkipVerify` | 布尔值 | 是否跳过 TLS 证书验证，生产环境应设为 `false` | `false` |
| `retryCount` | 整数 | 发送失败时的重试次数 | `3` |

## DKIM 签名配置

DKIM 签名可以提高邮件送达率，减少被标记为垃圾邮件的可能性。

```json
{
  "dkim": {
    "enabled": true,                             // 是否启用 DKIM 签名
    "domain": "example.com",                     // DKIM 域名
    "selector": "mail",                          // DKIM 选择器
    "privateKeyPath": "keys/example.com/mail.private", // DKIM 私钥路径
    "headersToSign": ["From", "To", "Subject", "Date", "Message-ID"], // 要签名的头部
    "signatureExpiry": 604800                    // 签名过期时间（秒）
  }
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `enabled` | 布尔值 | 是否启用 DKIM 签名 | `false` |
| `domain` | 字符串 | DKIM 域名，通常是发件人的域名 | 空，必须设置 |
| `selector` | 字符串 | DKIM 选择器名称 | `"mail"` |
| `privateKeyPath` | 字符串 | DKIM 私钥文件路径 | 空，必须设置 |
| `headersToSign` | 字符串数组 | 要签名的邮件头部字段 | 包含常用头部 |
| `signatureExpiry` | 整数 | 签名过期时间（秒）| `604800` (7 天) |

## 批处理与性能配置

这些配置项控制邮件的批量处理和性能相关参数。

```json
{
  "batchSize": 20,          // 每批发送的最大收件人数
  "batchDelay": 1000        // 批次间延迟（毫秒）
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `batchSize` | 整数 | 每批发送邮件的最大收件人数量 | `20` |
| `batchDelay` | 整数 | 批次间的延迟时间（毫秒） | `1000` |

## 健康检查配置

健康检查功能提供了监控服务状态的 HTTP 接口。

```json
{
  "enableHealthCheck": true,   // 是否启用健康检查
  "healthCheckPort": 8025      // 健康检查 HTTP 服务端口
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `enableHealthCheck` | 布尔值 | 是否启用健康检查 HTTP 服务 | `true` |
| `healthCheckPort` | 整数 | 健康检查 HTTP 服务端口 | `8025` |

## 速率限制配置

速率限制可以防止邮件发送过于频繁，影响邮件送达率。

```json
{
  "rateLimits": {
    "enabled": true,      // 是否启用速率限制
    "maxPerHour": 500,    // 每小时每发件人最大邮件数
    "maxPerDay": 2000     // 每天每发件人最大邮件数
  }
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `enabled` | 布尔值 | 是否启用速率限制 | `false` |
| `maxPerHour` | 整数 | 每小时每发件人可发送的最大邮件数 | `500` |
| `maxPerDay` | 整数 | 每天每发件人可发送的最大邮件数 | `2000` |

## 安全配置

安全配置控制服务器的安全相关选项。

```json
{
  "security": {
    "allowLocalOnly": true,   // 是否只允许本地连接
    "logAllEmails": true,     // 是否记录所有邮件内容
    "requireAuth": true       // 是否要求 SMTP 认证
  }
}
```

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `allowLocalOnly` | 布尔值 | 是否只允许本地连接 (127.0.0.1) | `true` |
| `logAllEmails` | 布尔值 | 是否记录所有邮件内容到日志 | `false` |
| `requireAuth` | 布尔值 | 是否要求 SMTP 认证 | `true` |

## 完整配置示例

以下是一个包含所有主要配置选项的完整示例：

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
  "forwardProviders": [
    {
      "host": "smtp.primary.com",
      "port": 587,
      "username": "user@primary.com",
      "password": "password1",
      "ssl": false,
      "priority": 0
    },
    {
      "host": "smtp.backup.com",
      "port": 587,
      "username": "user@backup.com",
      "password": "password2",
      "ssl": false,
      "priority": 1
    }
  ],
  
  "dkim": {
    "enabled": true,
    "domain": "example.com",
    "selector": "mail",
    "privateKeyPath": "keys/example.com/mail.private",
    "headersToSign": ["From", "To", "Subject", "Date", "Message-ID"],
    "signatureExpiry": 604800
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



## 配置最佳实践

1. **敏感信息安全**：避免在配置文件中存储明文密码，优先使用环境变量
2. **使用 DKIM 签名**：对于生产环境，强烈建议启用 DKIM 签名
3. **合理设置批量参数**：根据您的业务特点和目标邮件服务器的限制调整 `batchSize` 和 `batchDelay`
4. **启用多提供商故障转移**：配置至少两个 SMTP 服务提供商以提高可靠性
5. **定期备份配置**：保持配置文件的备份，特别是在修改前
6. **隔离不同环境**：开发、测试和生产环境应使用不同的配置文件

## 配置验证与调试

启动服务时，系统会自动验证配置并输出警告或错误信息。可以通过检查日志了解配置是否正确：

```bash
./mailer -config config.json
```

如需详细了解各配置项的使用，请参考相关功能的专题指南：
- [直接发送模式](/guides/direct-delivery)
- [DKIM 设置](/guides/dkim-setup)
- [多提供商故障转移](/guides/provider_failover)

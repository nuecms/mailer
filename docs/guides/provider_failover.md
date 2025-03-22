# 多 SMTP 提供商与故障转移

Go Mail Server 支持配置多个 SMTP 服务提供商，并在主要提供商不可用时自动切换到备用提供商。这种机制提高了系统的可靠性，确保即使某个 SMTP 服务出现问题，邮件仍然能够正常发送。

## 配置多个 SMTP 提供商

在 `config.json` 中使用 `forwardProviders` 数组配置多个 SMTP 提供商：

```json
{
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
    },
    {
      "host": "smtp.emergency.com",
      "port": 465,
      "username": "user@emergency.com",
      "password": "password3",
      "ssl": true,
      "priority": 2
    }
  ]
}
```

## 提供商优先级和故障转移机制

系统会按照以下规则处理邮件发送：

1. **优先级排序**：首先按照 `priority` 值对提供商进行排序（数字越小优先级越高）
2. **顺序尝试**：从优先级最高的提供商开始尝试发送邮件
3. **故障检测**：如果当前提供商发送失败（如连接超时、认证失败等），系统会自动尝试下一个提供商
4. **重试机制**：对每个提供商，系统会尝试最多 3 次发送，每次尝试之间的延迟时间呈指数增长
5. **失败处理**：如果所有提供商都发送失败，邮件会被保存到本地的 `emails/failed` 目录中，系统会在后续定期尝试重新发送

## 配置参数说明

需要注意的是，无论使用哪种方式配置SMTP提供商，`forwardSMTP`标志都作为总开关。如果设置为`false`，则无论是新的多提供商方式还是旧的单一提供商方式都不会生效。

每个 SMTP 提供商支持以下配置参数：

| 参数 | 类型 | 描述 | 默认值 |
|-----|-----|-----|-----|
| `host` | 字符串 | SMTP 服务器地址 | 必填 |
| `port` | 整数 | SMTP 服务器端口 | 必填 |
| `username` | 字符串 | SMTP 认证用户名 | 空（表示不需要认证） |
| `password` | 字符串 | SMTP 认证密码 | 空（表示不需要认证） |
| `ssl` | 布尔值 | 是否使用 SSL 连接（而不是 STARTTLS） | `false` |
| `priority` | 整数 | 提供商优先级，数字越小优先级越高 | 配置顺序 |

## 不同提供商配置示例

### Gmail

```json
{
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "ssl": false,
  "priority": 0
}
```

注意：Gmail 需要使用 App 密码而非常规密码。

### 阿里云企业邮箱

```json
{
  "host": "smtp.qiye.aliyun.com", 
  "port": 465,
  "username": "your-name@your-domain.com",
  "password": "your-password",
  "ssl": true,
  "priority": 1
}
```

### Amazon SES

```json
{
  "host": "email-smtp.us-east-1.amazonaws.com",
  "port": 587,
  "username": "YOUR_SES_SMTP_USERNAME",
  "password": "YOUR_SES_SMTP_PASSWORD",
  "ssl": false,
  "priority": 2
}
```

## 监控和日志

系统会记录每个提供商的发送尝试和结果。日志示例：

## 兼容性说明

为了保持向后兼容性，系统仍然支持旧版的单一提供商配置：

```json
{
  "forwardSMTP": true,              // 总开关，必须设为true才能启用转发
  "forwardHost": "smtp.example.com",
  "forwardPort": 587,
  "forwardUsername": "user@example.com",
  "forwardPassword": "password",
  "forwardSSL": false
}
```

如果同时存在旧版配置和多提供商配置，系统会优先使用多提供商配置。在新的部署中，我们建议直接使用多提供商配置方式，即使只有一个提供商：

```json
{
  "forwardSMTP": true,              // 总开关，必须设为true才能启用转发
  "forwardProviders": [
    {
      "host": "smtp.example.com",
      "port": 587,
      "username": "user@example.com",
      "password": "password",
      "ssl": false,
      "priority": 0
    }
  ]
}
```

而不使用旧版的单一提供商配置方式。旧版配置在未来版本中可能会被废弃。


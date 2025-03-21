# Go Mail Server 高级功能指南

本文档详细介绍了 Go Mail Server 的高级功能，包括异步处理、监控、批量发送和故障恢复等。

## 目录

1. [异步邮件处理](#异步邮件处理)
2. [批量邮件发送](#批量邮件发送)
3. [监控和健康检查](#监控和健康检查)
4. [速率限制](#速率限制)
5. [故障恢复](#故障恢复)
6. [Docker容器化](#docker容器化)
7. [配置说明](#配置说明)

## 异步邮件处理

Go Mail Server 采用异步处理模式，邮件接收后立即返回成功，在后台处理发送任务：

### 工作流程

1. SMTP服务接收邮件请求
2. 验证发件人和收件人
3. 将邮件放入内部队列
4. 立即返回成功响应给客户端
5. 工作协程从队列获取邮件并处理
6. 结果记录到日志和指标系统

### 优势

- **更快响应**：客户端不需等待邮件实际发送完成
- **更好的可用性**：即使转发SMTP服务暂时不可用，也能接收邮件
- **平滑处理峰值负载**：队列可以缓冲突发的邮件发送请求

### 自定义配置

```json
{
  "workerCount": 5,           // 工作协程数量
  "queueSize": 1000,          // 队列大小
  "processingTimeout": 300    // 单封邮件处理超时时间（秒）
}
```

## 批量邮件发送

当需要向大量收件人发送邮件时，系统会自动分批处理：

### 特性

- 自动将大量收件人分批处理
- 批次间自动添加延迟，避免触发接收方的速率限制
- 单个收件人失败不会影响整批发送
- 详细的批处理日志记录

### 配置选项

```json
{
  "batchSize": 20,          // 每批最大收件人数
  "batchDelay": 1000,       // 批次间延迟（毫秒）
  "continueOnError": true   // 部分收件人失败是否继续发送
}
```

### 使用场景

- 大规模通知邮件
- 电子报和营销邮件
- 任何需要发送给多个收件人的邮件

## 监控和健康检查

Go Mail Server 提供了全面的监控和健康检查API：

### 健康检查API

- **端点**: `http://localhost:8025/health`
- **方法**: GET
- **访问限制**: 仅本地访问
- **响应格式**: JSON

示例响应:
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

### 指标API

- **端点**: `http://localhost:8025/metrics`
- **方法**: GET
- **访问限制**: 仅本地访问
- **响应格式**: JSON

示例响应:
```json
{
  "total_emails": 1250,
  "successful_emails": 1247,
  "failed_emails": 3,
  "total_recipients": 5280,
  "avg_processing_time_ms": 235
}
```

### 集成监控系统

可以轻松集成到 Prometheus、Grafana、Zabbix 等监控系统中。

通过 curl 或其他HTTP客户端定期检查健康状态:
```bash
curl http://localhost:8025/health
curl http://localhost:8025/metrics
```

## 速率限制

为防止过度使用和避免被接收方标记为垃圾邮件，系统内置了多级速率限制：

### 实现的限制类型

- **每小时每发件人限制**：限制单个发件人每小时可发送的邮件数量
- **每天每发件人限制**：限制单个发件人每天可发送的邮件数量
- **全局速率限制**：限制整个系统每小时/每天可发送的邮件总数

### 配置选项

```json
{
  "rateLimits": {
    "enabled": true,
    "maxPerHour": 500,
    "maxPerDay": 2000,
    "globalMaxPerHour": 1000,
    "globalMaxPerDay": 5000
  }
}
```

### 超出限制时的行为

当发件人超出速率限制时，系统会：
1. 返回临时错误给SMTP客户端
2. 记录警告日志
3. 增加相应的指标计数
4. 客户端应当稍后重试发送

## 故障恢复

Go Mail Server 具有强大的故障恢复机制：

### 自动重试

- 临时故障（如网络问题）会自动重试
- 使用指数退避算法（exponential backoff）减轻对目标服务器的压力
- 可配置的重试次数和间隔

### 失败邮件存储

- 所有发送失败的邮件都保存到 `emails/failed` 目录
- JSON格式保存完整的邮件元数据和内容
- 定期尝试重新发送失败的邮件

### 手动恢复

管理员可以手动触发失败邮件的重新发送：
```bash
curl -X POST http://localhost:8025/admin/retry-failed
```

## Docker容器化

Go Mail Server 提供了完整的Docker支持：

### 使用Docker Compose启动

```bash
docker-compose up -d
```

### 容器健康检查

容器包含内置的健康检查，确保服务正常运行。

### 数据持久化

使用Docker卷保存配置和邮件数据：
- `/app/config.json`：配置文件
- `/app/emails`：邮件存储目录

### 环境变量支持

可以通过环境变量覆盖配置：
```
SMTP_HOST=0.0.0.0
SMTP_PORT=25
DEFAULT_USERNAME=admin
DEFAULT_PASSWORD=password
FORWARD_SMTP=true
FORWARD_HOST=smtp.example.com
...
```

## 配置说明

完整的配置选项说明：

| 分类 | 选项 | 类型 | 默认值 | 说明 |
|------|------|------|------|------|
| **基础配置** |
| | smtpHost | string | "127.0.0.1" | SMTP服务器监听地址 |
| | smtpPort | int | 25 | SMTP服务器监听端口 |
| | defaultUsername | string | "" | SMTP认证用户名 |
| | defaultPassword | string | "" | SMTP认证密码 |
| **转发设置** |
| | forwardSMTP | bool | false | 是否启用转发 |
| | forwardHost | string | "" | 转发SMTP服务器地址 |
| | forwardPort | int | 587 | 转发SMTP服务器端口 |
| | forwardUsername | string | "" | 转发SMTP认证用户名 |
| | forwardPassword | string | "" | 转发SMTP认证密码 |
| | forwardSSL | bool | false | 是否使用SSL连接转发服务器 |
| **批量发送** |
| | batchSize | int | 20 | 每批最大收件人数量 |
| | batchDelay | int | 1000 | 批次间延迟（毫秒） |
| | continueOnError | bool | true | 部分失败时是否继续发送 |
| **监控设置** |
| | enableHealthCheck | bool | true | 是否启用健康检查API |
| | healthCheckPort | int | 8025 | 健康检查HTTP服务端口 |
| **速率限制** |
| | rateLimits.enabled | bool | false | 是否启用速率限制 |
| | rateLimits.maxPerHour | int | 500 | 每小时每发件人最大邮件数 |
| | rateLimits.maxPerDay | int | 2000 | 每天每发件人最大邮件数 |
| | rateLimits.globalMaxPerHour | int | 1000 | 每小时系统总邮件数限制 |
| | rateLimits.globalMaxPerDay | int | 5000 | 每天系统总邮件数限制 |
| **安全设置** |
| | security.allowLocalOnly | bool | true | 是否只允许本地连接 |
| | security.logAllEmails | bool | true | 是否记录所有邮件内容 |
| | security.requireAuth | bool | true | 是否要求认证 |

通过以上高级功能，Go Mail Server 可以满足从小型开发环境到大型生产环境的各种邮件发送需求，同时保持高可靠性和安全性。

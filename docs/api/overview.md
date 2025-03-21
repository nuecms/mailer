# API 概述

Go Mail Server 提供了一组 HTTP API，用于监控服务状态、获取性能指标和执行管理操作。这些 API 默认只接受来自本地的连接，以确保安全性。

## API 端点列表

| 端点 | 方法 | 描述 |
| --- | --- | --- |
| `/health` | GET | 获取服务健康状态 |
| `/metrics` | GET | 获取性能指标 |
| `/admin/retry-failed` | POST | 触发重新处理失败邮件 |

## 认证和安全

所有 API 端点默认只接受来自本地地址 (`127.0.0.1`, `localhost`) 的连接。这是一种安全措施，防止未授权的远程访问。

如果需要从远程访问这些 API 端点，建议使用以下方法之一：

1. SSH 隧道
2. 反向代理 (如 Nginx) 并添加认证
3. Cloudflare Tunnel (详见 [Cloudflare Tunnel 部署指南](/guides/cloudflare-tunnel))

## 响应格式

所有 API 响应均以 JSON 格式返回。标准响应结构如下：

```json
{
  "status": "ok",             // 状态：ok 或 error
  "timestamp": "ISO8601时间戳", // 响应生成时间
  ... 其他特定于端点的字段 ...
}
```

## 错误处理

当发生错误时，API 返回适当的 HTTP 状态码和 JSON 格式的错误详情：

```json
{
  "status": "error",
  "error": "错误描述",
  "timestamp": "ISO8601时间戳"
}
```

## 使用示例

使用 `curl` 访问健康检查端点：

```bash
curl http://localhost:8025/health
```

使用 `curl` 触发失败邮件重试：

```bash
curl -X POST http://localhost:8025/admin/retry-failed
```

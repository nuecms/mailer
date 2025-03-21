# 健康检查 API

健康检查 API 提供了对服务当前健康状态的详细信息，包括磁盘空间使用情况、队列状态和处理统计数据。

## 端点信息

- **URL**: `/health`
- **方法**: GET
- **权限**: 仅本地连接

## 响应格式

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

## 响应字段说明

| 字段 | 类型 | 描述 |
| --- | --- | --- |
| `status` | string | 服务状态，`ok` 表示正常 |
| `timestamp` | string | ISO8601 格式的时间戳 |
| `details` | object | 详细状态信息 |
| `details.disk_space_available_mb` | number | 可用磁盘空间 (MB) |
| `details.queued_emails` | number | 队列中等待处理的邮件数 |
| `details.failed_emails` | number | 失败的邮件数 |
| `details.total_emails_processed` | number | 已处理的邮件总数 |
| `details.success_rate` | number | 发送成功率 (百分比) |

## 使用示例

### 使用 curl 获取健康状态

```bash
curl http://localhost:8025/health
```

### 使用 Python 获取健康状态

```python
import requests
import json

response = requests.get('http://localhost:8025/health')
data = response.json()

print(f"服务状态: {data['status']}")
print(f"队列中邮件: {data['details']['queued_emails']}")
print(f"失败邮件: {data['details']['failed_emails']}")
print(f"成功率: {data['details']['success_rate']}%")
```

## 监控集成

您可以将健康检查 API 集成到各种监控系统中：

### Prometheus + Grafana

通过自定义导出器将健康检查数据转换为 Prometheus 指标。

### 自定义脚本监控

创建定期检查健康状态的脚本，并在发现问题时发送告警：

```bash
#!/bin/bash
response=$(curl -s http://localhost:8025/health)
status=$(echo $response | jq -r '.status')

if [ "$status" != "ok" ]; then
  echo "邮件服务异常!"
  exit 1
fi

# 检查失败邮件数
failed=$(echo $response | jq -r '.details.failed_emails')
if [ "$failed" -gt 100 ]; then
  echo "警告: 失败邮件数过多 ($failed)"
  # 发送告警...
fi
```

## 故障排查

如果无法访问健康检查 API，请检查：

1. 服务是否正在运行
2. 健康检查端口是否正确配置 (默认 8025)
3. 是否存在防火墙阻止连接
4. 日志文件中是否有相关错误信息

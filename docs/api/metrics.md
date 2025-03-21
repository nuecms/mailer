# 指标 API

指标 API 提供了关于邮件服务性能和活动的详细统计数据，用于监控和调优服务。

## 端点信息

- **URL**: `/metrics`
- **方法**: GET
- **权限**: 仅本地连接

## 响应格式

```json
{
  "total_emails": 1500,
  "successful_emails": 1450,
  "failed_emails": 50,
  "total_recipients": 2500,
  "avg_processing_time_ms": 240
}
```

## 响应字段说明

| 字段 | 类型 | 描述 |
| --- | --- | --- |
| `total_emails` | number | 处理的邮件总数 |
| `successful_emails` | number | 成功发送的邮件数 |
| `failed_emails` | number | 发送失败的邮件数 |
| `total_recipients` | number | 收件人总数 (一封邮件可能有多个收件人) |
| `avg_processing_time_ms` | number | 平均处理时间 (毫秒) |

## 使用示例

### 使用 curl 获取指标

```bash
curl http://localhost:8025/metrics
```

### 使用 Node.js 获取指标

```javascript
const http = require('http');

http.get('http://localhost:8025/metrics', (res) => {
  let data = '';
  
  res.on('data', (chunk) => {
    data += chunk;
  });
  
  res.on('end', () => {
    const metrics = JSON.parse(data);
    console.log('邮件总数:', metrics.total_emails);
    console.log('成功率:', (metrics.successful_emails / metrics.total_emails * 100).toFixed(2) + '%');
    console.log('平均处理时间:', metrics.avg_processing_time_ms, 'ms');
  });
}).on('error', (err) => {
  console.error('获取指标失败:', err.message);
});
```

## 指标监控与分析

### 性能监控

关注 `avg_processing_time_ms` 值的变化趋势。如果这个值开始增加，可能表明存在性能问题或外部 SMTP 服务器响应缓慢。

### 成功率监控

计算成功率 = `successful_emails / total_emails * 100%`。如果成功率低于预期，应该检查：

1. 外部 SMTP 服务器状态
2. 网络连接问题
3. 身份验证配置是否正确
4. 收件人地址是否有效

### 资源使用监控

结合 `/health` 端点提供的磁盘空间信息，监控系统资源使用情况。

## 与 Prometheus 集成

要将指标导出到 Prometheus，可以创建一个简单的导出器脚本：

```python
from prometheus_client import start_http_server, Gauge
import requests
import time

# 创建 Prometheus 指标
total_emails = Gauge('mailer_total_emails', 'Total emails processed')
successful_emails = Gauge('mailer_successful_emails', 'Successfully sent emails')
failed_emails = Gauge('mailer_failed_emails', 'Failed emails')
processing_time = Gauge('mailer_avg_processing_time_ms', 'Average processing time in ms')

def get_metrics():
    try:
        response = requests.get('http://localhost:8025/metrics')
        data = response.json()
        
        total_emails.set(data['total_emails'])
        successful_emails.set(data['successful_emails'])
        failed_emails.set(data['failed_emails'])
        processing_time.set(data['avg_processing_time_ms'])
    except Exception as e:
        print(f"Error fetching metrics: {e}")

if __name__ == '__main__':
    # 在端口 9091 启动 metrics 服务器
    start_http_server(9091)
    
    while True:
        get_metrics()
        time.sleep(15)  # 每15秒更新一次
```

## 指标数据保留

当前的指标数据保存在内存中，服务重启后会重置。如果需要长期保留历史指标数据，建议：

1. 定期收集指标并存储到外部数据库或时间序列数据库
2. 使用 Prometheus + Grafana 等工具进行长期存储和可视化
3. 设置定期导出指标数据到 CSV 或 JSON 文件

# Go Mail Server 优化指南

本文档提供了优化仅用于发送邮件且只允许本地连接的邮件服务的建议。

## 目录

1. [性能优化](#性能优化)
2. [安全加固](#安全加固)
3. [监控与告警](#监控与告警)
4. [高可用配置](#高可用配置)
5. [故障恢复](#故障恢复)

## 性能优化

### 批量处理

当需要向大量收件人发送相同邮件时，使用批处理功能可以显著提高性能：

```go
// 已在代码中实现批处理逻辑，每批最多20个收件人
```

配置文件中可添加以下设置：
```json
{
  "batchSize": 20,
  "batchDelay": 1000  // 每批之间的延迟（毫秒）
}
```

### 连接池

为了避免频繁建立和关闭连接，可以实现SMTP连接池：

```go
// 连接池示例代码
type SMTPClient struct {
    client *smtp.Client
    lastUsed time.Time
}

var smtpPool = make(chan *SMTPClient, 10) // 最多10个连接
```

### 异步发送

将邮件发送改为异步模式，立即返回结果并在后台处理发送：

```go
// 创建邮件队列
var mailQueue = make(chan MailJob, 1000)

// 启动多个工作协程处理队列
for i := 0; i < 5; i++ {
    go func() {
        for job := range mailQueue {
            // 处理邮件发送
        }
    }()
}

// 将邮件放入队列而不是直接发送
mailQueue <- MailJob{From: from, To: to, Data: data}
```

## 安全加固

### 限制连接

确保服务器只接受来自本地的连接：

```go
func isLocalConnection(addr net.Addr) bool {
    // 当前已实现，只接受来自本地的连接
    ipStr, _, _ := net.SplitHostPort(addr.String())
    ip := net.ParseIP(ipStr)
    return ip.IsLoopback() || ipStr == "::1" || strings.HasPrefix(ipStr, "127.")
}
```

### 速率限制

实现速率限制防止过度使用：

```go
// 一个简单的基于时间窗口的速率限制器
type RateLimiter struct {
    limit     int
    interval  time.Duration
    requests  map[string][]time.Time
    mu        sync.Mutex
}

func (r *RateLimiter) Allow(key string) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-r.interval)
    
    // 清理过期记录
    var recent []time.Time
    for _, t := range r.requests[key] {
        if t.After(cutoff) {
            recent = append(recent, t)
        }
    }
    
    r.requests[key] = recent
    
    // 检查是否超过限制
    if len(recent) >= r.limit {
        return false
    }
    
    // 记录新请求
    r.requests[key] = append(r.requests[key], now)
    return true
}
```

### 日志增强

添加更详细的日志记录以便审计和故障排除：

```go
// 为每封邮件生成唯一ID
mailID := generateUUID()
log.Printf("[%s] 收到发送请求: 从 %s 到 %s (%d 个收件人)", 
    mailID, from, strings.Join(to[:min(3, len(to))], ","), len(to))

// 记录详细结果
log.Printf("[%s] 邮件发送完成, 耗时: %v, 结果: %v", 
    mailID, time.Since(startTime), err == nil)
```

## 监控与告警

### 健康检查端点

添加一个HTTP服务以提供健康检查和状态监控：

```go
// 启动HTTP服务器提供监控端点
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    
    // 只允许本地连接
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    ip := net.ParseIP(host)
    if !ip.IsLoopback() {
        w.WriteHeader(http.StatusForbidden)
        return
    }
    
    health := systemHealthCheck()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
})

go func() {
    if err := http.ListenAndServe("127.0.0.1:8025", nil); err != nil {
        log.Printf("健康检查HTTP服务启动失败: %v", err)
    }
}()
```

### 指标收集

收集关键指标以监控系统性能：

```go
type Metrics struct {
    totalEmails      int64
    successfulEmails int64
    failedEmails     int64
    totalRecipients  int64
    processingTime   time.Duration
    mu               sync.Mutex
}

func (m *Metrics) RecordSuccess(recipients int, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.totalEmails++
    m.successfulEmails++
    m.totalRecipients += int64(recipients)
    m.processingTime += duration
}

func (m *Metrics) RecordFailure(recipients int, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.totalEmails++
    m.failedEmails++
    m.totalRecipients += int64(recipients)
    m.processingTime += duration
}
```

## 高可用配置

### 多实例部署

在多台服务器上部署多个实例，通过负载均衡器分发流量：

```
+----------------+      +-----------------+
| 应用服务器 1    |      | 邮件服务实例 1  |
+----------------+  /-> +-----------------+
                    |
+----------------+  |   +-----------------+
| 应用服务器 2    |--+-> | 邮件服务实例 2  |
+----------------+  |   +-----------------+
                    |
+----------------+  \-> +-----------------+
| 应用服务器 3    |      | 邮件服务实例 3  |
+----------------+      +-----------------+
```

### 失败转移

配置多个SMTP服务提供商，当一个失败时自动切换到备用提供商：

```go
// 配置多个转发服务器
var forwardProviders = []ForwardConfig{
    {Host: "smtp.primary.com", Port: 587, Username: "user1", Password: "pass1"},
    {Host: "smtp.backup.com", Port: 587, Username: "user2", Password: "pass2"},
}

// 尝试每个提供商直到成功
for _, provider := range forwardProviders {
    if err := sendWithProvider(provider, from, to, data); err == nil {
        return nil
    } else {
        log.Printf("使用提供商 %s 发送失败: %v", provider.Host, err)
    }
}
```

## 故障恢复

### 本地队列

将所有邮件保存到本地队列，定期尝试重新发送失败的邮件：

```go
// 定期检查并尝试重新发送失败的邮件
func processFailedEmails() {
    dir := "emails"
    files, _ := os.ReadDir(dir)
    
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".failed") {
            // 读取失败的邮件并尝试重新发送
            data, err := os.ReadFile(filepath.Join(dir, file.Name()))
            if err != nil {
                continue
            }
            
            // 解析邮件元数据并重试发送
            // ...
        }
    }
}

// 每小时运行一次重试处理
go func() {
    ticker := time.NewTicker(time.Hour)
    for range ticker.C {
        processFailedEmails()
    }
}()
```

### 监控告警

设置监控系统，当检测到异常时发送告警：

```go
// 检查队列积压
func checkQueueBacklog() {
    dir := "emails"
    files, _ := os.ReadDir(dir)
    
    count := 0
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".failed") {
            count++
        }
    }
    
    if count > 100 {
        sendAlert(fmt.Sprintf("邮件队列积压: %d 封邮件等待处理", count))
    }
}

// 发送告警
func sendAlert(message string) {
    // 可以通过多种方式发送告警：
    // 1. 写入日志
    log.Printf("告警: %s", message)
    
    // 2. 调用外部API
    // http.Post("https://alert-service/api/alert", "application/json", ...)
    
    // 3. 向管理员发送邮件（如果邮件系统仍然工作）
    // ...
}
```

## 总结

通过实施上述优化，您的邮件发送服务将变得更加高效、可靠和安全。根据您的具体需求和资源，可以选择性地实施这些优化措施。

记住，对于仅用于发送邮件且只允许本地连接的服务来说，安全性和可靠性是最重要的考虑因素。确保服务的连接限制严格执行，并实现适当的监控和告警机制，以便及时发现和解决问题。

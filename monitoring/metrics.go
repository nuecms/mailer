package monitoring

import (
	"sync"
	"time"

	"github.com/nuecms/mailer/config"
)

// Metrics 存储服务性能指标
type Metrics struct {
	TotalEmails      int64
	SuccessfulEmails int64
	FailedEmails     int64
	TotalRecipients  int64
	ProcessingTime   time.Duration
	Requests         map[string][]time.Time
	Mu               sync.Mutex
}

// NewMetrics 创建一个新的指标实例
func NewMetrics() *Metrics {
	return &Metrics{
		Requests: make(map[string][]time.Time),
	}
}

// RecordSuccess 记录成功发送的邮件
func (m *Metrics) RecordSuccess(recipients int, duration time.Duration) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.TotalEmails++
	m.SuccessfulEmails++
	m.TotalRecipients += int64(recipients)
	m.ProcessingTime += duration
}

// RecordFailure 记录发送失败的邮件
func (m *Metrics) RecordFailure(recipients int, duration time.Duration) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.TotalEmails++
	m.FailedEmails++
	m.TotalRecipients += int64(recipients)
	m.ProcessingTime += duration
}

// CheckRateLimit 检查速率限制
func (m *Metrics) CheckRateLimit(from string, config *config.Config) bool {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	now := time.Now()

	// 清理过期的小时级别记录
	hourCutoff := now.Add(-time.Hour)
	var recentHour []time.Time
	for _, t := range m.Requests[from+"_hour"] {
		if t.After(hourCutoff) {
			recentHour = append(recentHour, t)
		}
	}
	m.Requests[from+"_hour"] = recentHour

	// 清理过期的天级别记录
	dayCutoff := now.Add(-24 * time.Hour)
	var recentDay []time.Time
	for _, t := range m.Requests[from+"_day"] {
		if t.After(dayCutoff) {
			recentDay = append(recentDay, t)
		}
	}
	m.Requests[from+"_day"] = recentDay

	// 检查是否超过限制
	if len(recentHour) >= config.RateLimits.MaxPerHour {
		return false
	}
	if len(recentDay) >= config.RateLimits.MaxPerDay {
		return false
	}

	// 记录新请求
	m.Requests[from+"_hour"] = append(m.Requests[from+"_hour"], now)
	m.Requests[from+"_day"] = append(m.Requests[from+"_day"], now)
	return true
}

// GetMetricsData 获取指标数据
func (m *Metrics) GetMetricsData() map[string]interface{} {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	result := map[string]interface{}{
		"total_emails":      m.TotalEmails,
		"successful_emails": m.SuccessfulEmails,
		"failed_emails":     m.FailedEmails,
		"total_recipients":  m.TotalRecipients,
		"avg_processing_time_ms": int64(0),
	}

	if m.TotalEmails > 0 {
		result["avg_processing_time_ms"] = int64(m.ProcessingTime/time.Millisecond) / m.TotalEmails
	}

	return result
}

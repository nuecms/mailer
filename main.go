package main

import (
	"flag"
	"log"
	"time"

	"github.com/nuecms/mailer/config"
	"github.com/nuecms/mailer/mail"
	"github.com/nuecms/mailer/monitoring"
	"github.com/nuecms/mailer/server"
	"github.com/nuecms/mailer/utils"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if (err != nil) {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 检查转发设置
	config.CheckForwardingConfig(cfg)

	// 创建指标收集器
	metrics := monitoring.NewMetrics()

	// 创建邮件队列
	mailQueue := make(chan mail.MailJob, 1000)

	// 启动工作协程处理队列
	for i := 0; i < 5; i++ {
		go processMailQueue(i+1, cfg, metrics, mailQueue)
	}

	// 启动健康检查HTTP服务
	if cfg.EnableHealthCheck {
		go monitoring.StartHealthCheckServer(cfg.HealthCheckPort, metrics)
	}

	// 启动定期任务
	go startPeriodicTasks()

	// 启动SMTP服务器
	if err := server.SetupAndRunSMTPServer(cfg, metrics, mailQueue); err != nil {
		log.Fatalf("SMTP服务器启动失败: %v", err)
	}
}

// 处理邮件队列的工作协程
func processMailQueue(workerID int, cfg *config.Config, metrics *monitoring.Metrics, mailQueue chan mail.MailJob) {
	log.Printf("启动邮件处理工作协程 #%d", workerID)
	for job := range mailQueue {
		startTime := time.Now()
		log.Printf("[%s] 工作协程 #%d 处理邮件: 从 %s 到 %s",
			job.ID, workerID, job.From, utils.SummarizeRecipients(job.To))

		var err error
		if cfg.ForwardSMTP && cfg.ForwardHost != "" {
			err = mail.ForwardMail(cfg, job.From, job.To, job.Data)
		} else {
			err = mail.SaveMailLocally(job.From, job.To, job.Data)
		}

		if err != nil {
			log.Printf("[%s] 邮件处理失败: %v", job.ID, err)
			metrics.RecordFailure(len(job.To), time.Since(startTime))

			// 保存失败的邮件
			if saveErr := mail.SaveFailedMail(job); saveErr != nil {
				log.Printf("[%s] 保存失败邮件失败: %v", job.ID, saveErr)
			}
		} else {
			log.Printf("[%s] 邮件处理成功, 耗时: %v", job.ID, time.Since(startTime))
			metrics.RecordSuccess(len(job.To), time.Since(startTime))
		}
	}
}

// 启动定期任务
func startPeriodicTasks() {
	// 每小时尝试重新发送失败的邮件
	go func() {
		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
			mail.ProcessFailedEmails()
		}
	}()

	// 每5分钟检查队列积压
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			utils.CheckQueueBacklog()
		}
	}()
}
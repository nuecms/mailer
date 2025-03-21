package mail

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuecms/mailer/utils"
)

// SaveMailLocally 保存邮件到本地文件系统
func SaveMailLocally(from string, to []string, data []byte) error {
	// 创建邮件目录（如果不存在）
	mailDir := "emails"
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		return fmt.Errorf("创建邮件目录失败: %v", err)
	}

	// 生成唯一的文件名
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s/%s_%s.eml", mailDir, timestamp, strings.Replace(from, "@", "_at_", -1))

	// 创建邮件文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建邮件文件失败: %v", err)
	}
	defer file.Close()

	// 写入元数据
	fmt.Fprintf(file, "From: %s\n", from)
	fmt.Fprintf(file, "To: %s\n", strings.Join(to, ", "))
	fmt.Fprintf(file, "Date: %s\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(file, "\n")

	// 写入原始邮件数据
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("写入邮件数据失败: %v", err)
	}

	log.Printf("邮件已保存到: %s", filename)
	return nil
}

// SaveFailedMail 保存失败的邮件以便稍后重试
func SaveFailedMail(job MailJob) error {
	dir := "emails/failed"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建失败邮件目录失败: %v", err)
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s.json", job.ID))

	// 将作业信息保存为JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("序列化作业数据失败: %v", err)
	}

	if err := os.WriteFile(filename, jobData, 0644); err != nil {
		return fmt.Errorf("写入失败邮件文件失败: %v", err)
	}

	log.Printf("[%s] 失败邮件已保存到: %s", job.ID, filename)
	return nil
}

// ProcessFailedEmails 处理失败的邮件
func ProcessFailedEmails() {
	log.Printf("开始处理失败邮件...")
	dir := "emails/failed"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return // 目录不存在，没有失败邮件
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("读取失败邮件目录失败: %v", err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(dir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("读取失败邮件文件失败: %v", err)
				continue
			}

			var job MailJob
			if err := json.Unmarshal(data, &job); err != nil {
				log.Printf("解析失败邮件JSON失败: %v", err)
				continue
			}

			log.Printf("[%s] 尝试重新发送失败邮件: 从 %s 到 %s",
				job.ID, job.From, utils.SummarizeRecipients(job.To))

			// 尝试重新发送
			if err := ForwardMail(nil, job.From, job.To, job.Data); err != nil {
				log.Printf("[%s] 重新发送失败: %v", job.ID, err)
			} else {
				log.Printf("[%s] 重新发送成功", job.ID)
				// 删除成功发送的失败邮件文件
				if err := os.Remove(filePath); err != nil {
					log.Printf("[%s] 删除失败邮件文件失败: %v", job.ID, err)
				}
			}
		}
	}
	log.Printf("失败邮件处理完成")
}

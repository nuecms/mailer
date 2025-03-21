# Python 使用示例

本文档提供了使用 Python 连接 Go Mail Server 发送邮件的示例代码。

## 基本邮件发送

```python
import smtplib
from email.message import EmailMessage

def send_email(subject, body, to_emails):
    """
    使用 Go Mail Server 发送邮件
    
    Args:
        subject: 邮件主题
        body: 邮件正文
        to_emails: 收件人列表，可以是单个字符串或列表
    """
    # 准备邮件内容
    msg = EmailMessage()
    msg.set_content(body)
    msg['Subject'] = subject
    msg['From'] = 'sender@example.com'
    
    # 处理收件人
    if isinstance(to_emails, str):
        to_emails = [to_emails]
    msg['To'] = ', '.join(to_emails)
    
    # 连接到 Go Mail Server
    smtp_host = 'localhost'  # Go Mail Server 地址
    smtp_port = 25           # SMTP 端口
    smtp_user = 'noreply@example.com'  # 配置文件中的 defaultUsername
    smtp_pass = 'your-password'        # 配置文件中的 defaultPassword
    
    try:
        # 创建 SMTP 连接
        s = smtplib.SMTP(smtp_host, smtp_port)
        
        # 开始 TLS 加密 (如果服务器支持)
        try:
            s.starttls()
        except smtplib.SMTPNotSupportedError:
            # 如果不支持 TLS，继续使用明文
            pass
        
        # 认证
        s.login(smtp_user, smtp_pass)
        
        # 发送邮件
        s.send_message(msg)
        print(f"邮件已发送到: {', '.join(to_emails)}")
        
        # 关闭连接
        s.quit()
        return True
    
    except Exception as e:
        print(f"发送邮件时出错: {e}")
        return False

# 使用示例
if __name__ == "__main__":
    send_email(
        subject="测试邮件",
        body="这是一封测试邮件，由 Go Mail Server 发送。",
        to_emails=["recipient@example.com", "another@example.com"]
    )
```

## 发送 HTML 邮件

```python
import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText

def send_html_email(subject, html_body, to_emails):
    """
    使用 Go Mail Server 发送 HTML 邮件
    
    Args:
        subject: 邮件主题
        html_body: HTML 格式的邮件正文
        to_emails: 收件人列表，可以是单个字符串或列表
    """
    # 准备邮件内容
    msg = MIMEMultipart('alternative')
    msg['Subject'] = subject
    msg['From'] = 'sender@example.com'
    
    # 处理收件人
    if isinstance(to_emails, str):
        to_emails = [to_emails]
    msg['To'] = ', '.join(to_emails)
    
    # 添加 HTML 内容
    msg.attach(MIMEText(html_body, 'html'))
    
    # 连接到 Go Mail Server
    smtp_host = 'localhost'
    smtp_port = 25
    smtp_user = 'noreply@example.com'
    smtp_pass = 'your-password'
    
    try:
        s = smtplib.SMTP(smtp_host, smtp_port)
        try:
            s.starttls()
        except smtplib.SMTPNotSupportedError:
            pass
        s.login(smtp_user, smtp_pass)
        s.send_message(msg)
        s.quit()
        return True
    except Exception as e:
        print(f"发送邮件时出错: {e}")
        return False

# 使用示例
if __name__ == "__main__":
    html = """
    <html>
      <body>
        <h1>测试 HTML 邮件</h1>
        <p>这是一封 <b>HTML</b> 格式的测试邮件，由 Go Mail Server 发送。</p>
        <p>您可以添加图片、链接和其他 HTML 元素。</p>
        <a href="https://example.com">访问我们的网站</a>
      </body>
    </html>
    """
    send_html_email(
        subject="HTML 测试邮件",
        html_body=html,
        to_emails="recipient@example.com"
    )
```

## 发送带附件的邮件

```python
import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.mime.application import MIMEApplication
import os

def send_email_with_attachment(subject, body, to_emails, attachment_path):
    """
    使用 Go Mail Server 发送带附件的邮件
    
    Args:
        subject: 邮件主题
        body: 邮件正文
        to_emails: 收件人列表，可以是单个字符串或列表
        attachment_path: 附件文件路径
    """
    # 准备邮件内容
    msg = MIMEMultipart()
    msg['Subject'] = subject
    msg['From'] = 'sender@example.com'
    
    # 处理收件人
    if isinstance(to_emails, str):
        to_emails = [to_emails]
    msg['To'] = ', '.join(to_emails)
    
    # 添加邮件正文
    msg.attach(MIMEText(body))
    
    # 添加附件
    with open(attachment_path, 'rb') as file:
        attachment = MIMEApplication(file.read(), Name=os.path.basename(attachment_path))
    
    attachment['Content-Disposition'] = f'attachment; filename="{os.path.basename(attachment_path)}"'
    msg.attach(attachment)
    
    # 连接到 Go Mail Server
    smtp_host = 'localhost'
    smtp_port = 25
    smtp_user = 'noreply@example.com'
    smtp_pass = 'your-password'
    
    try:
        s = smtplib.SMTP(smtp_host, smtp_port)
        try:
            s.starttls()
        except smtplib.SMTPNotSupportedError:
            pass
        s.login(smtp_user, smtp_pass)
        s.send_message(msg)
        s.quit()
        return True
    except Exception as e:
        print(f"发送邮件时出错: {e}")
        return False

# 使用示例
if __name__ == "__main__":
    send_email_with_attachment(
        subject="带附件的测试邮件",
        body="这是一封带附件的测试邮件，由 Go Mail Server 发送。",
        to_emails="recipient@example.com",
        attachment_path="document.pdf"
    )
```

## 使用外部 Tunnel 访问

如果您使用 Cloudflare Tunnel 配置了外部访问，可以按照以下方式连接：

```python
def send_email_via_tunnel(subject, body, to_emails):
    """
    通过 Cloudflare Tunnel 使用 Go Mail Server 发送邮件
    """
    # 准备邮件内容
    msg = EmailMessage()
    msg.set_content(body)
    msg['Subject'] = subject
    msg['From'] = 'sender@example.com'
    
    if isinstance(to_emails, str):
        to_emails = [to_emails]
    msg['To'] = ', '.join(to_emails)
    
    # 使用 Tunnel 域名连接
    smtp_host = 'smtp.yourdomain.com'  # 替换为您的 Tunnel 域名
    smtp_port = 25
    smtp_user = 'noreply@example.com'
    smtp_pass = 'your-password'
    
    try:
        s = smtplib.SMTP(smtp_host, smtp_port)
        try:
            s.starttls()
        except smtplib.SMTPNotSupportedError:
            pass
        s.login(smtp_user, smtp_pass)
        s.send_message(msg)
        s.quit()
        return True
    except Exception as e:
        print(f"发送邮件时出错: {e}")
        return False
```

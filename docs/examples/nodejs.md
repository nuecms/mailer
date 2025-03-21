# Node.js 使用示例

本文档提供了使用 Node.js 连接 Go Mail Server 发送邮件的示例代码。

## 使用 Nodemailer 发送邮件

[Nodemailer](https://nodemailer.com/) 是 Node.js 环境中最流行的邮件发送库。

### 安装 Nodemailer

```bash
npm install nodemailer
```

### 基本邮件发送

```javascript
const nodemailer = require('nodemailer');

// 创建 SMTP 传输对象
const transporter = nodemailer.createTransport({
  host: 'localhost',     // Go Mail Server 地址
  port: 25,              // SMTP 端口
  auth: {
    user: 'noreply@example.com',  // 配置文件中的 defaultUsername
    pass: 'your-password'         // 配置文件中的 defaultPassword
  }
});

// 定义邮件选项
const mailOptions = {
  from: 'sender@example.com',            // 发件人
  to: 'recipient@example.com',           // 收件人
  subject: 'Node.js 测试邮件',            // 主题
  text: '这是一封使用 Node.js 和 Nodemailer 发送的测试邮件。'  // 正文
};

// 发送邮件
transporter.sendMail(mailOptions, (error, info) => {
  if (error) {
    console.error('发送邮件失败:', error);
  } else {
    console.log('邮件已发送:', info.response);
  }
});
```

### 使用 Promise/async-await

```javascript
const nodemailer = require('nodemailer');

async function sendEmail(to, subject, body) {
  // 创建传输对象
  const transporter = nodemailer.createTransport({
    host: 'localhost',
    port: 25,
    auth: {
      user: 'noreply@example.com',
      pass: 'your-password'
    }
  });
  
  // 定义邮件选项
  const mailOptions = {
    from: 'sender@example.com',
    to: to,
    subject: subject,
    text: body
  };
  
  try {
    // 发送邮件
    const info = await transporter.sendMail(mailOptions);
    console.log('邮件已发送:', info.response);
    return true;
  } catch (error) {
    console.error('发送邮件失败:', error);
    return false;
  }
}

// 使用示例
sendEmail(
  'recipient@example.com',
  'async/await 测试邮件',
  '这是使用 async/await 发送的测试邮件。'
).then(result => {
  console.log('操作完成，结果:', result);
});
```

### 发送 HTML 邮件

```javascript
const nodemailer = require('nodemailer');

async function sendHtmlEmail(to, subject, htmlContent) {
  const transporter = nodemailer.createTransport({
    host: 'localhost',
    port: 25,
    auth: {
      user: 'noreply@example.com',
      pass: 'your-password'
    }
  });
  
  const mailOptions = {
    from: 'sender@example.com',
    to: to,
    subject: subject,
    html: htmlContent  // 使用 HTML 内容
  };
  
  try {
    const info = await transporter.sendMail(mailOptions);
    console.log('HTML 邮件已发送:', info.response);
    return true;
  } catch (error) {
    console.error('发送 HTML 邮件失败:', error);
    return false;
  }
}

// 使用示例
const htmlContent = `
  <h1>HTML 邮件测试</h1>
  <p>这是一封 <strong>HTML</strong> 格式的邮件。</p>
  <ul>
    <li>支持 HTML 标签</li>
    <li>可以包含样式</li>
    <li>支持图片和链接</li>
  </ul>
  <p>访问我们的网站: <a href="https://example.com">Example.com</a></p>
`;

sendHtmlEmail(
  'recipient@example.com',
  'HTML 格式邮件测试',
  htmlContent
);
```

### 发送带附件的邮件

```javascript
const nodemailer = require('nodemailer');
const fs = require('fs');
const path = require('path');

async function sendEmailWithAttachment(to, subject, body, attachmentPath) {
  const transporter = nodemailer.createTransport({
    host: 'localhost',
    port: 25,
    auth: {
      user: 'noreply@example.com',
      pass: 'your-password'
    }
  });
  
  // 准备附件
  const attachment = {
    filename: path.basename(attachmentPath),
    content: fs.createReadStream(attachmentPath)
  };
  
  const mailOptions = {
    from: 'sender@example.com',
    to: to,
    subject: subject,
    text: body,
    attachments: [attachment]  // 可以添加多个附件
  };
  
  try {
    const info = await transporter.sendMail(mailOptions);
    console.log('带附件的邮件已发送:', info.response);
    return true;
  } catch (error) {
    console.error('发送带附件的邮件失败:', error);
    return false;
  }
}

// 使用示例
sendEmailWithAttachment(
  'recipient@example.com',
  '带附件的邮件',
  '请查看附件中的文档。',
  './documents/sample.pdf'
);
```

## 发送批量邮件

```javascript
const nodemailer = require('nodemailer');

async function sendBulkEmails(recipients, subject, body) {
  const transporter = nodemailer.createTransport({
    host: 'localhost',
    port: 25,
    auth: {
      user: 'noreply@example.com',
      pass: 'your-password'
    }
  });
  
  // 使用 Promise.all 并行发送多封邮件
  const results = await Promise.all(
    recipients.map(async (recipient) => {
      const mailOptions = {
        from: 'sender@example.com',
        to: recipient,
        subject: subject,
        text: body
      };
      
      try {
        const info = await transporter.sendMail(mailOptions);
        console.log(`邮件已发送到 ${recipient}:`, info.response);
        return { recipient, success: true };
      } catch (error) {
        console.error(`发送邮件到 ${recipient} 失败:`, error);
        return { recipient, success: false, error: error.message };
      }
    })
  );
  
  // 统计发送结果
  const successful = results.filter(r => r.success).length;
  console.log(`批量邮件发送完成: ${successful}/${recipients.length} 成功`);
  
  return results;
}

// 使用示例
const recipients = [
  'user1@example.com',
  'user2@example.com',
  'user3@example.com',
  'user4@example.com'
];

sendBulkEmails(
  recipients,
  '批量邮件测试',
  '这是一封测试批量发送功能的邮件。'
);
```

## 使用外部 Tunnel 访问

如果您使用 Cloudflare Tunnel 配置了外部访问：

```javascript
const nodemailer = require('nodemailer');

// 使用 Cloudflare Tunnel 域名
const transporter = nodemailer.createTransport({
  host: 'smtp.yourdomain.com',  // 替换为您的 Tunnel 域名
  port: 25,
  auth: {
    user: 'noreply@example.com',
    pass: 'your-password'
  }
});

const mailOptions = {
  from: 'sender@example.com',
  to: 'recipient@example.com',
  subject: 'Tunnel 测试邮件',
  text: '这封邮件通过 Cloudflare Tunnel 发送。'
};

transporter.sendMail(mailOptions, (error, info) => {
  if (error) {
    console.error('发送邮件失败:', error);
  } else {
    console.log('邮件已发送:', info.response);
  }
});
```

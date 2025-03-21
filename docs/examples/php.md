# PHP 使用示例

本文档提供了使用 PHP 连接 Go Mail Server 发送邮件的示例代码。

## 使用 mail() 函数

PHP 内置的 `mail()` 函数可以通过配置 `php.ini` 来使用本地 SMTP 服务器。

### 配置 php.ini

要使用 Go Mail Server，您需要修改 PHP 配置文件 `php.ini`：

```ini
[mail function]
; 将 SMTP 设置为 Go Mail Server 地址
SMTP = localhost
smtp_port = 25

; 如果 Go Mail Server 需要认证
sendmail_from = noreply@example.com
```

### 基本邮件发送

```php
<?php
// 收件人
$to = 'recipient@example.com';

// 主题
$subject = 'PHP 测试邮件';

// 邮件内容
$message = '这是一封使用 PHP mail() 函数发送的测试邮件。';

// 邮件头部
$headers = 'From: sender@example.com' . "\r\n" .
    'Reply-To: sender@example.com' . "\r\n" .
    'X-Mailer: PHP/' . phpversion();

// 发送邮件
$success = mail($to, $subject, $message, $headers);

if ($success) {
    echo "邮件已成功发送到 $to";
} else {
    echo "发送邮件失败";
}
?>
```

## 使用 PHPMailer 库

[PHPMailer](https://github.com/PHPMailer/PHPMailer) 是 PHP 中最流行的邮件发送库，它提供了更多高级功能。

### 安装 PHPMailer

使用 Composer 安装：

```bash
composer require phpmailer/phpmailer
```

### 基本邮件发送

```php
<?php
// 引入 Composer 自动加载器
require 'vendor/autoload.php';

// 导入 PHPMailer 类
use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

// 创建 PHPMailer 实例
$mail = new PHPMailer(true); // true 启用异常

try {
    // 服务器设置
    $mail->isSMTP();
    $mail->Host       = 'localhost'; // Go Mail Server 地址
    $mail->SMTPAuth   = true;
    $mail->Username   = 'noreply@example.com'; // 配置文件中的 defaultUsername
    $mail->Password   = 'your-password';       // 配置文件中的 defaultPassword
    $mail->Port       = 25;
    
    // 收件人
    $mail->setFrom('sender@example.com', '发件人名称');
    $mail->addAddress('recipient@example.com', '收件人名称');
    $mail->addReplyTo('reply@example.com', '回复地址');
    
    // 邮件内容
    $mail->isHTML(false);
    $mail->Subject = 'PHPMailer 测试邮件';
    $mail->Body    = '这是使用 PHPMailer 发送的测试邮件。';
    
    // 发送邮件
    $mail->send();
    echo '邮件已成功发送';
} catch (Exception $e) {
    echo "发送邮件失败: {$mail->ErrorInfo}";
}
?>
```

### 发送 HTML 邮件

```php
<?php
require 'vendor/autoload.php';

use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

$mail = new PHPMailer(true);

try {
    // 服务器设置
    $mail->isSMTP();
    $mail->Host       = 'localhost';
    $mail->SMTPAuth   = true;
    $mail->Username   = 'noreply@example.com';
    $mail->Password   = 'your-password';
    $mail->Port       = 25;
    
    // 收件人
    $mail->setFrom('sender@example.com', '发件人名称');
    $mail->addAddress('recipient@example.com', '收件人名称');
    
    // HTML 内容
    $mail->isHTML(true);
    $mail->Subject = 'HTML 格式邮件测试';
    $mail->Body    = '
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; }
            .header { color: #2C3E50; font-size: 24px; }
            .content { margin: 20px 0; line-height: 1.5; }
            .footer { color: #7F8C8D; font-size: 12px; }
        </style>
    </head>
    <body>
        <div class="header">HTML 邮件测试</div>
        <div class="content">
            <p>这是一封 <strong>HTML</strong> 格式的邮件，支持各种 HTML 标签和样式。</p>
            <p>您可以添加链接：<a href="https://example.com">访问网站</a></p>
        </div>
        <div class="footer">此邮件由 Go Mail Server 发送</div>
    </body>
    </html>';
    
    // 纯文本替代内容（用于不支持 HTML 的邮件客户端）
    $mail->AltBody = '这是一封 HTML 格式的邮件，如果您的邮件客户端不支持 HTML，将显示此文本。';
    
    // 发送邮件
    $mail->send();
    echo 'HTML 邮件已成功发送';
} catch (Exception $e) {
    echo "发送邮件失败: {$mail->ErrorInfo}";
}
?>
```

### 发送带附件的邮件

```php
<?php
require 'vendor/autoload.php';

use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

$mail = new PHPMailer(true);

try {
    // 服务器设置
    $mail->isSMTP();
    $mail->Host       = 'localhost';
    $mail->SMTPAuth   = true;
    $mail->Username   = 'noreply@example.com';
    $mail->Password   = 'your-password';
    $mail->Port       = 25;
    
    // 收件人
    $mail->setFrom('sender@example.com', '发件人名称');
    $mail->addAddress('recipient@example.com', '收件人名称');
    
    // 邮件内容
    $mail->isHTML(true);
    $mail->Subject = '带附件的邮件';
    $mail->Body    = '<p>这是一封带附件的邮件，请查看附件。</p>';
    $mail->AltBody = '这是一封带附件的邮件，请查看附件。';
    
    // 添加附件
    $mail->addAttachment('/path/to/document.pdf', '文档.pdf');    // 可选文件名
    $mail->addAttachment('/path/to/image.jpg', '图片.jpg');
    
    // 发送邮件
    $mail->send();
    echo '带附件的邮件已成功发送';
} catch (Exception $e) {
    echo "发送邮件失败: {$mail->ErrorInfo}";
}
?>
```

## 使用 Cloudflare Tunnel 访问

如果您使用 Cloudflare Tunnel 配置了外部访问：

```php
<?php
require 'vendor/autoload.php';

use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

$mail = new PHPMailer(true);

try {
    // 使用 Tunnel 域名
    $mail->isSMTP();
    $mail->Host       = 'smtp.yourdomain.com'; // 替换为您的 Tunnel 域名
    $mail->SMTPAuth   = true;
    $mail->Username   = 'noreply@example.com';
    $mail->Password   = 'your-password';
    $mail->Port       = 25;
    
    // 收件人
    $mail->setFrom('sender@example.com', '发件人名称');
    $mail->addAddress('recipient@example.com', '收件人名称');
    
    // 邮件内容
    $mail->Subject = 'Tunnel 测试邮件';
    $mail->Body    = '这封邮件通过 Cloudflare Tunnel 发送。';
    
    // 发送邮件
    $mail->send();
    echo '通过 Tunnel 发送邮件成功';
} catch (Exception $e) {
    echo "发送邮件失败: {$mail->ErrorInfo}";
}
?>
```

# DKIM 签名设置指南

DKIM (DomainKeys Identified Mail) 是一种用于电子邮件身份认证的标准方法，它允许邮件接收方验证邮件确实是由声称的发送域发送并且邮件内容在传输过程中没有被篡改。本指南将帮助您为 Go Mail Server 设置和配置 DKIM 签名。

## 目录

- [DKIM 签名设置指南](#dkim-签名设置指南)
  - [目录](#目录)
  - [DKIM 工作原理](#dkim-工作原理)
  - [设置步骤](#设置步骤)
    - [1. 生成 DKIM 密钥](#1-生成-dkim-密钥)
      - [使用 OpenSSL 生成 RSA 密钥对](#使用-openssl-生成-rsa-密钥对)
      - [将公钥转换为 DNS 记录格式](#将公钥转换为-dns-记录格式)
    - [2. 配置 DNS 记录](#2-配置-dns-记录)

## DKIM 工作原理

DKIM 的工作流程如下：

1. **生成密钥对**：域名所有者生成一对公钥和私钥
2. **配置 DNS**：公钥被发布到域名的 DNS 记录中
3. **签名邮件**：使用私钥对外发邮件进行签名
4. **验证签名**：接收方获取公钥并验证签名的有效性

## 设置步骤

### 1. 生成 DKIM 密钥

#### 使用 OpenSSL 生成 RSA 密钥对

```bash
# 创建密钥目录
mkdir -p /opt/mailer/keys/example.com

# 生成私钥
openssl genrsa -out /opt/mailer/keys/example.com/mail.private 2048

# 从私钥中提取公钥
openssl rsa -in /opt/mailer/keys/example.com/mail.private -pubout -outform PEM -out /opt/mailer/keys/example.com/mail.public
```

#### 将公钥转换为 DNS 记录格式

```bash
# 从公钥提取用于 DNS 记录的部分（移除头部、尾部和换行符）
cat /opt/mailer/keys/example.com/mail.public | grep -v '^-' | tr -d '\n'
```

### 2. 配置 DNS 记录

您需要在您的 DNS 管理面板中添加一条 TXT 记录：


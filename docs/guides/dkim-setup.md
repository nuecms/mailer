# DKIM 设置指南

DKIM (DomainKeys Identified Mail) 是一种电子邮件验证技术，可以让接收方验证邮件确实来自您声称的域名，并且在传输过程中没有被篡改。配置 DKIM 可以提高您的邮件送达率，减少被标记为垃圾邮件的几率。

## DKIM 工作原理

DKIM 使用公钥加密技术：

1. 发件方服务器使用私钥对邮件的特定部分进行签名
2. 接收方服务器通过 DNS 查询获取发件方域名的公钥
3. 接收方使用公钥验证签名的有效性
4. 如果验证成功，证明邮件确实由该域名发送且未被篡改

## Go Mail Server 中的 DKIM 配置

### 自动配置（推荐）

最简单的方法是使用我们提供的自动配置脚本：

```bash
# 下载并运行配置脚本
chmod +x setup_tunnel.sh
./setup_tunnel.sh
```

该脚本会：
1. 生成 DKIM 密钥对（公钥和私钥）
2. 提供需要添加到 DNS 的 TXT 记录
3. 更新配置文件以启用 DKIM

### 手动配置

如果您想手动配置 DKIM，请按照以下步骤操作：

#### 1. 生成 DKIM 密钥对

```bash
# 创建目录
mkdir -p keys/example.com

# 生成私钥
openssl genrsa -out keys/example.com/mail.private 2048

# 从私钥中提取公钥
openssl rsa -in keys/example.com/mail.private -pubout -outform PEM -out keys/example.com/mail.public

# 转换为 DNS 记录格式
cat keys/example.com/mail.public | grep -v '^-' | tr -d '\n' > keys/example.com/mail.txt
```

#### 2. 添加 DNS 记录

在您的域名 DNS 控制面板中，添加以下 TXT 记录：


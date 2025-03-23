#!/bin/bash

# 设置颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== Go Mail Server - DKIM 配置工具 =====${NC}"
echo -e "${YELLOW}注意: 本脚本仅配置 DKIM 签名，不再支持 Cloudflare Tunnel 配置${NC}"
echo -e "${YELLOW}Cloudflare WARP 客户端默认会阻止端口 25 上的出站 SMTP 流量，使其不适用于 SMTP 服务器${NC}"
echo -e "${YELLOW}详情请参阅: https://developers.cloudflare.com/cloudflare-one/faq/troubleshooting/#i-cannot-send-emails-on-port-25${NC}"
echo

# 获取配置文件路径
CONFIG_FILE="config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    if [ -f "config.example.json" ]; then
        echo -e "${YELLOW}配置文件不存在，将基于示例配置创建${NC}"
        cp config.example.json config.json
    else
        echo -e "${RED}错误: 找不到配置文件，请确保 config.json 或 config.example.json 存在${NC}"
        exit 1
    fi
fi

# 从配置中提取域名
if command -v jq &> /dev/null; then
    DEFAULT_USERNAME=$(jq -r '.defaultUsername' "$CONFIG_FILE")
    if [[ "$DEFAULT_USERNAME" == *"@"* ]]; then
        DOMAIN=$(echo "$DEFAULT_USERNAME" | cut -d@ -f2)
        echo -e "${GREEN}从配置中提取到域名: $DOMAIN${NC}"
    else
        echo -e "${YELLOW}配置中的用户名没有包含域名部分，请手动指定域名${NC}"
        read -p "请输入您的域名 (例如 example.com): " DOMAIN
    fi
else
    echo -e "${YELLOW}未检测到 jq 工具，无法自动提取域名${NC}"
    read -p "请输入您的域名 (例如 example.com): " DOMAIN
fi

# 确认域名正确性
echo -e "将使用 ${YELLOW}$DOMAIN${NC} 作为邮件域名，请确认是否正确"
read -p "继续? (y/n): " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    read -p "请输入正确的域名: " DOMAIN
fi

# 设置 DKIM 密钥
echo -e "\n${GREEN}=== 设置 DKIM 密钥 ===${NC}"
mkdir -p "keys/$DOMAIN"

SELECTOR="mail"
PRIVATE_KEY="keys/$DOMAIN/${SELECTOR}.private"
PUBLIC_KEY="keys/$DOMAIN/${SELECTOR}.public"
DNS_KEY="keys/$DOMAIN/${SELECTOR}.txt"

# 检查是否已存在密钥
if [ -f "$PRIVATE_KEY" ]; then
    echo -e "${YELLOW}发现已存在的 DKIM 密钥，是否重新生成?${NC}"
    read -p "重新生成密钥? (y/n): " REGENERATE
    if [[ "$REGENERATE" != "y" && "$REGENERATE" != "Y" ]]; then
        echo "将使用现有密钥继续"
    else
        rm -f "$PRIVATE_KEY" "$PUBLIC_KEY" "$DNS_KEY"
    fi
fi

# 生成新密钥（如果需要）
if [ ! -f "$PRIVATE_KEY" ]; then
    echo "生成 DKIM 密钥对..."
    
    # 检查 openssl 是否可用
    if ! command -v openssl &> /dev/null; then
        echo -e "${RED}未检测到 openssl，无法生成 DKIM 密钥${NC}"
        echo "请安装 openssl 后重试"
        exit 1
    fi
    
    # 生成私钥
    openssl genrsa -out "$PRIVATE_KEY" 2048
    
    # 从私钥中提取公钥
    openssl rsa -in "$PRIVATE_KEY" -pubout -outform PEM -out "$PUBLIC_KEY"
    
    # 转换为 DNS 记录格式
    cat "$PUBLIC_KEY" | grep -v '^-' | tr -d '\n' > "$DNS_KEY"
    
    echo -e "${GREEN}DKIM 密钥生成完成!${NC}"
fi

# 创建 DNS 记录
PUBLIC_KEY_VALUE=$(cat "$PUBLIC_KEY" | grep -v '^-' | tr -d '\n')
DKIM_DNS_RECORD="${SELECTOR}._domainkey.${DOMAIN}. IN TXT \"v=DKIM1; k=rsa; p=${PUBLIC_KEY_VALUE}\""

echo -e "${YELLOW}请添加以下 DKIM DNS 记录:${NC}"
echo "${SELECTOR}._domainkey.${DOMAIN}. IN TXT \"v=DKIM1; k=rsa; p=${PUBLIC_KEY_VALUE}\""

# 更新配置文件添加 DKIM 设置
echo -e "\n${GREEN}=== 更新 DKIM 配置 ===${NC}"
echo "你的 DKIM 配置信息:"
echo "  - 域名:   $DOMAIN"
echo "  - 选择器: $SELECTOR"
echo "  - 私钥:   $PRIVATE_KEY"

# 提示用户更新配置文件
echo -e "\n${YELLOW}请确保在 config.json 中添加或更新以下 DKIM 配置:${NC}"
echo -e "\"dkim\": {
  \"enabled\": true,
  \"domain\": \"$DOMAIN\",
  \"selector\": \"$SELECTOR\",
  \"privateKeyPath\": \"$PRIVATE_KEY\",
  \"headersToSign\": [\"From\", \"To\", \"Subject\", \"Date\", \"Message-ID\"],
  \"signatureExpiry\": 604800
}"

# 创建 DNS 验证脚本
cat > verify_dns.sh << EOF
#!/bin/bash
# 验证 DNS 记录是否正确配置

DOMAIN="$DOMAIN"
SELECTOR="$SELECTOR"

echo "正在验证 MX 记录..."
dig MX \$DOMAIN +short

echo -e "\n正在验证 SPF 记录..."
dig TXT \$DOMAIN +short | grep "spf1"

echo -e "\n正在验证 DKIM 记录..."
dig TXT \${SELECTOR}._domainkey.\${DOMAIN} +short

echo -e "\n正在验证 DMARC 记录..."
dig TXT _dmarc.\${DOMAIN} +short
EOF
chmod +x verify_dns.sh

# 创建 DNS 配置指南
cat << EOF > dns_config_guide.txt
# DNS 配置指南

为了提高邮件送达率和减少被标记为垃圾邮件的几率，请在你的 DNS 提供商控制面板中添加以下记录:

## 必要记录

1. **DKIM 记录** - 用于验证邮件的真实性
   ${SELECTOR}._domainkey.${DOMAIN}. IN TXT "v=DKIM1; k=rsa; p=${PUBLIC_KEY_VALUE}"

## 强烈推荐添加的记录

2. **SPF 记录** - 指定允许发送邮件的服务器
   ${DOMAIN}.               IN TXT "v=spf1 ip4:YOUR_SERVER_IP ~all"
   
   * 将 YOUR_SERVER_IP 替换为您服务器的实际 IP 地址

3. **DMARC 记录** - 邮件认证、报告和一致性策略
   _dmarc.${DOMAIN}.        IN TXT "v=DMARC1; p=none; rua=mailto:admin@${DOMAIN}"

## 配置完成后验证

使用以下命令验证你的 DNS 记录是否正确配置:
./verify_dns.sh
EOF

echo -e "${GREEN}已创建 DNS 配置指南: dns_config_guide.txt${NC}"
echo -e "${GREEN}已创建 DNS 验证脚本: verify_dns.sh${NC}"
echo -e "${GREEN}DKIM 配置设置完成!${NC}"
echo -e "${YELLOW}请确保按照 dns_config_guide.txt 文件中的说明配置您的 DNS 记录${NC}"
echo -e "${YELLOW}配置完成后，使用 ./verify_dns.sh 验证 DNS 记录是否正确配置${NC}"
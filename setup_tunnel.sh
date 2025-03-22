#!/bin/bash

# 设置颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== Go Mail Server - Cloudflare Tunnel & DKIM 配置工具 =====${NC}"

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

# Step 1: 设置 DKIM 密钥
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
PUBLIC_KEY_VALUE=$(cat "$DNS_KEY")
DKIM_DNS_RECORD="${SELECTOR}._domainkey.${DOMAIN}. IN TXT \"v=DKIM1; k=rsa; p=${PUBLIC_KEY_VALUE}\""

echo -e "${YELLOW}请添加以下 DKIM DNS 记录:${NC}"
echo "$DKIM_DNS_RECORD"

# 更新配置文件添加 DKIM 设置（可选，视配置文件结构而定）
echo -e "\n${GREEN}=== 更新 DKIM 配置 ===${NC}"
echo "你的 DKIM 配置信息:"
echo "  - 域名:   $DOMAIN"
echo "  - 选择器: $SELECTOR"
echo "  - 私钥:   $PRIVATE_KEY"
echo 
echo "在您的代码中集成 DKIM 签名时，请使用上述信息"

# Step 2: 设置 Cloudflare Tunnel
echo -e "\n${GREEN}=== 设置 Cloudflare Tunnel ===${NC}"

# 检查 cloudflared 是否已安装
if ! command -v cloudflared &> /dev/null; then
    echo -e "${YELLOW}未检测到 cloudflared，正在尝试安装...${NC}"
    
    # 根据系统类型安装 cloudflared
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
        sudo dpkg -i cloudflared.deb
        rm cloudflared.deb
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install cloudflare/cloudflare/cloudflared
    else
        echo -e "${RED}不支持的操作系统。请手动安装 cloudflared: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation${NC}"
        exit 1
    fi
fi

# 检查是否已登录 Cloudflare
echo "检查 Cloudflare 认证状态..."
if ! cloudflared tunnel list &> /dev/null; then
    echo -e "${YELLOW}需要登录 Cloudflare，请在浏览器中完成认证...${NC}"
    cloudflared login
fi

# 设置 Tunnel 名称
TUNNEL_NAME="mailer-tunnel"
MAIL_PREFIX="mail"
SMTP_PREFIX="smtp"

# 检查是否存在名为 $TUNNEL_NAME 的 tunnel
echo "检查是否存在名为 $TUNNEL_NAME 的 tunnel..."
if cloudflared tunnel list | grep -q "$TUNNEL_NAME"; then
    echo "找到现有的 tunnel: $TUNNEL_NAME"
    TUNNEL_ID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | awk '{print $1}')
else
    echo "创建新的 tunnel: $TUNNEL_NAME"
    TUNNEL_ID=$(cloudflared tunnel create $TUNNEL_NAME | grep -oP 'Created tunnel \K[a-z0-9-]+')
    
    echo "将 DNS 记录关联到 tunnel..."
    # 配置 DNS 记录
    cloudflared tunnel route dns $TUNNEL_ID "${MAIL_PREFIX}.${DOMAIN}"
    cloudflared tunnel route dns $TUNNEL_ID "${SMTP_PREFIX}.${DOMAIN}"
fi

echo "创建 tunnel 配置文件..."
mkdir -p ~/.cloudflared

# 创建配置文件
cat > ~/.cloudflared/config-$TUNNEL_NAME.yml << EOF
tunnel: $TUNNEL_ID
credentials-file: ~/.cloudflared/${TUNNEL_ID}.json

ingress:
  - hostname: ${MAIL_PREFIX}.${DOMAIN}
    service: http://localhost:8025
  - hostname: ${SMTP_PREFIX}.${DOMAIN}
    service: tcp://localhost:25
  - service: http_status:404
EOF

echo -e "${GREEN}Tunnel 配置完成! 配置文件: ~/.cloudflared/config-$TUNNEL_NAME.yml${NC}"

# 创建 DNS 配置指南
echo -e "${YELLOW}请在你的 DNS 提供商控制面板中添加以下记录:${NC}"
cat << EOF > dns_config_guide.txt
# Cloudflare DNS 配置指南

# MX 记录 - 指定负责接收域名邮件的服务器
${DOMAIN}.               IN    MX    10    ${MAIL_PREFIX}.${DOMAIN}.

# SPF 记录 - 指定允许发送邮件的服务器
${DOMAIN}.               IN    TXT    "v=spf1 include:_spf.${DOMAIN} ~all"
_spf.${DOMAIN}.          IN    TXT    "v=spf1 include:${SMTP_PREFIX}.${DOMAIN} ~all"

# DKIM 记录 - 用于验证邮件的真实性
${SELECTOR}._domainkey.${DOMAIN}. IN TXT "${DKIM_DNS_RECORD}"

# DMARC 记录 - 邮件认证、报告和一致性策略
_dmarc.${DOMAIN}.        IN    TXT    "v=DMARC1; p=none; rua=mailto:admin@${DOMAIN}"

# 注意: Cloudflare Tunnel 会自动为以下记录创建 CNAME:
# - ${MAIL_PREFIX}.${DOMAIN}
# - ${SMTP_PREFIX}.${DOMAIN}
EOF

echo -e "${GREEN}DNS 配置指南已创建: dns_config_guide.txt${NC}"

# 创建服务文件
echo "创建 Tunnel 服务文件..."
cat > cloudflared-tunnel.service << EOF
[Unit]
Description=Cloudflare Tunnel for Mail Server
After=network.target

[Service]
Type=simple
User=$(whoami)
ExecStart=/usr/local/bin/cloudflared tunnel --config ~/.cloudflared/config-$TUNNEL_NAME.yml run $TUNNEL_NAME
Restart=on-failure
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

[Install]
WantedBy=multi-user.target
EOF

echo -e "${GREEN}Tunnel 服务文件已创建: cloudflared-tunnel.service${NC}"
echo -e "${YELLOW}如需安装为系统服务:${NC}"
echo "sudo mv cloudflared-tunnel.service /etc/systemd/system/"
echo "sudo systemctl daemon-reload"
echo "sudo systemctl enable cloudflared-tunnel"
echo "sudo systemctl start cloudflared-tunnel"

# 创建启动脚本
cat > start_tunnel.sh << EOF
#!/bin/bash
# 启动 Cloudflare Tunnel

# 设置颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

TUNNEL_NAME="$TUNNEL_NAME"
CONFIG_PATH="\$HOME/.cloudflared/config-$TUNNEL_NAME.yml"

echo -e "\${GREEN}启动 Cloudflare Tunnel: \${TUNNEL_NAME}\${NC}"
echo -e "使用配置文件: \${CONFIG_PATH}"

# 正确的命令格式: cloudflared tunnel --config <配置文件> run <隧道名称>
cloudflared tunnel --config "\${CONFIG_PATH}" run "\${TUNNEL_NAME}"

# 检查是否启动成功
if [ \$? -ne 0 ]; then
    echo -e "\${RED}Tunnel 启动失败，请检查错误信息\${NC}"
    echo -e "\${YELLOW}请检查以下几点:\${NC}"
    echo "1. 配置文件路径是否正确"
    echo "2. 是否已登录 Cloudflare" 
    echo "3. 隧道名称是否正确"
    echo
    echo -e "\${YELLOW}可以尝试直接使用隧道ID运行:\${NC}"
    echo "cloudflared tunnel --config \${CONFIG_PATH} run $TUNNEL_ID"
    exit 1
fi
EOF
chmod +x start_tunnel.sh

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

echo -e "\n验证 Cloudflare Tunnel DNS 记录..."
dig ${MAIL_PREFIX}.\${DOMAIN} +short
dig ${SMTP_PREFIX}.\${DOMAIN} +short
EOF
chmod +x verify_dns.sh

# 更新 Go Mail Server 配置指南
cat > tunnel_mail_config.md << EOF
# Cloudflare Tunnel 与 DKIM 配置指南

## 当前配置信息

- **域名**: $DOMAIN
- **Tunnel ID**: $TUNNEL_ID
- **DKIM 选择器**: $SELECTOR

## 配置步骤

1. **DNS 记录**

   请在您的 DNS 提供商控制面板中添加以下记录:
   - MX 记录: \`${DOMAIN}. IN MX 10 ${MAIL_PREFIX}.${DOMAIN}.\`
   - SPF 记录: \`${DOMAIN}. IN TXT "v=spf1 include:_spf.${DOMAIN} ~all"\`
   - SPF 子域记录: \`_spf.${DOMAIN}. IN TXT "v=spf1 include:${SMTP_PREFIX}.${DOMAIN} ~all"\`
   - DKIM 记录: \`${SELECTOR}._domainkey.${DOMAIN}. IN TXT "v=DKIM1; k=rsa; p=..."\`
   - DMARC 记录: \`_dmarc.${DOMAIN}. IN TXT "v=DMARC1; p=none; rua=mailto:admin@${DOMAIN}"\`

   详细信息请参考 \`dns_config_guide.txt\` 文件。

2. **修改 Go Mail Server 配置**

   更新 \`config.json\` 中的以下设置:

   \`\`\`json
   {
     "security": {
       "allowLocalOnly": false  // 允许来自 Cloudflare Tunnel 的连接
     },
     "dkim": {
       "enabled": true,
       "domain": "${DOMAIN}",
       "selector": "${SELECTOR}",
       "privateKeyPath": "${PRIVATE_KEY}"
     }
   }
   \`\`\`

3. **启动服务**

   运行 Cloudflare Tunnel:
   \`\`\`
   ./start_tunnel.sh
   \`\`\`

   或手动运行:
   \`\`\`
   cloudflared tunnel --config ~/.cloudflared/config-$TUNNEL_NAME.yml run $TUNNEL_NAME
   \`\`\`

   启动 Go Mail Server:
   \`\`\`
   ./mailer -config config.json
   \`\`\`

4. **验证配置**

   运行 DNS 验证脚本:
   \`\`\`
   ./verify_dns.sh
   \`\`\`

## 使用方法

将您的应用程序 SMTP 服务器配置为:

- **SMTP 服务器**: smtp.${DOMAIN}
- **端口**: 25
- **用户名**: (配置文件中的 defaultUsername)
- **密码**: (配置文件中的 defaultPassword)

## 故障排除

如果邮件无法发送，请检查以下几点:

1. 确认 Cloudflare Tunnel 正在运行 (\`./start_tunnel.sh\`)
2. 确认 Go Mail Server 正在运行
3. 确认所有 DNS 记录已正确配置
4. 查看 Go Mail Server 日志中的错误信息

### 常见 Tunnel 问题

1. **错误: "flag provided but not defined: -config"**:
   - 正确的命令格式是 \`cloudflared tunnel --config <配置文件> run <隧道名称>\`
   - 而不是 \`cloudflared tunnel run --config <配置文件> <隧道名称>\`

2. **无法连接到 Tunnel**:
   - 确保 Cloudflare 账户已正确授权: \`cloudflared login\`
   - 检查隧道配置文件是否正确

3. **DNS 记录未生效**:
   - DNS 记录可能需要时间传播（通常几分钟到几小时）
   - 使用 \`./verify_dns.sh\` 检查记录是否正确设置
EOF

echo -e "${YELLOW}是否立即启动 Tunnel? [y/N]${NC}"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "启动 Tunnel..."
    # 使用正确的命令格式
    cloudflared tunnel --config ~/.cloudflared/config-$TUNNEL_NAME.yml run $TUNNEL_NAME
else
    echo -e "${GREEN}可以使用以下命令启动 Tunnel:${NC}"
    echo "./start_tunnel.sh"
    echo "或者直接运行:"
    echo "cloudflared tunnel --config ~/.cloudflared/config-$TUNNEL_NAME.yml run $TUNNEL_NAME"
fi

echo -e "${GREEN}配置和设置脚本执行完成!${NC}"
echo "请参考 tunnel_mail_config.md 文件了解完整的配置指南"
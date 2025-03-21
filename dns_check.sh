#!/bin/bash

# 彩色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== DNS & Email 配置检查工具 =====${NC}"

# 获取域名
read -p "请输入要检查的域名: " DOMAIN

echo -e "\n${YELLOW}正在检查 MX 记录...${NC}"
dig MX $DOMAIN +short

echo -e "\n${YELLOW}正在检查 SPF 记录...${NC}"
dig TXT $DOMAIN +short | grep "spf1"

echo -e "\n${YELLOW}正在检查 DKIM 记录...${NC}"
dig TXT mail._domainkey.$DOMAIN +short

echo -e "\n${YELLOW}正在检查 DMARC 记录...${NC}"
dig TXT _dmarc.$DOMAIN +short

echo -e "\n${YELLOW}正在检查邮件服务器连接...${NC}"
MAIL_SERVER=$(dig MX $DOMAIN +short | sort -n | head -1 | awk '{print $2}')
if [ -n "$MAIL_SERVER" ]; then
    echo "尝试连接到邮件服务器: $MAIL_SERVER"
    nc -z -w5 $(echo $MAIL_SERVER | sed 's/.$//' ) 25
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}连接成功!${NC}"
    else
        echo -e "${RED}连接失败!${NC}"
    fi
else
    echo -e "${RED}未找到邮件服务器!${NC}"
fi

echo -e "\n${YELLOW}验证发送配置...${NC}"
echo "如需测试发送功能，请运行:"
echo "  echo \"Subject: 测试邮件\" | curl --ssl-reqd --url \"smtps://smtp.$DOMAIN:465\" --user \"your-username:your-password\" --mail-from \"sender@$DOMAIN\" --mail-rcpt \"recipient@example.com\" --upload-file -"

echo -e "\n${GREEN}检查完成!${NC}"
echo "如果您看到任何错误或缺失记录，请参考 dns_config_guide.txt 进行配置"
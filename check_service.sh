#!/bin/bash

# 颜色设置
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== Go Mail Server 服务检查工具 =====${NC}"

# 检查配置文件
if [ ! -f config.json ]; then
    echo -e "${RED}错误: 配置文件 config.json 不存在${NC}"
    echo "是否需要从模板创建? (y/n)"
    read answer
    if [ "$answer" == "y" ]; then
        cp config.example.json config.json
        echo -e "${YELLOW}已创建 config.json，请编辑配置信息${NC}"
        exit 1
    else
        echo "请创建配置文件后再运行服务"
        exit 1
    fi
fi

# 检查端口占用
echo -e "\n${YELLOW}检查端口占用情况...${NC}"

# 从配置中读取端口，如果不能读取则使用默认值
SMTP_PORT=$(grep -o '"smtpPort": *[0-9]*' config.json | awk '{print $2}')
SMTP_PORT=${SMTP_PORT:-25}

HEALTH_PORT=$(grep -o '"healthCheckPort": *[0-9]*' config.json | awk '{print $2}')
HEALTH_PORT=${HEALTH_PORT:-8025}

# 检查SMTP端口
if nc -z localhost $SMTP_PORT 2>/dev/null; then
    echo -e "${RED}SMTP端口 $SMTP_PORT 已被占用${NC}"
    lsof -i :$SMTP_PORT
    echo "请修改配置文件中的 smtpPort 设置"
else
    echo -e "${GREEN}SMTP端口 $SMTP_PORT 可用${NC}"
fi

# 检查健康检查端口
if nc -z localhost $HEALTH_PORT 2>/dev/null; then
    echo -e "${YELLOW}健康检查端口 $HEALTH_PORT 已被占用，服务会自动尝试其他端口${NC}"
else
    echo -e "${GREEN}健康检查端口 $HEALTH_PORT 可用${NC}"
fi

# 检查服务是否已运行
PID=$(pgrep -f "mailer -config")
if [ ! -z "$PID" ]; then
    echo -e "${YELLOW}Go Mail Server 已经在运行，PID: $PID${NC}"
    echo "使用以下命令停止服务:"
    echo "  kill $PID"
fi

# 检查目录权限
echo -e "\n${YELLOW}检查目录权限...${NC}"
mkdir -p emails/failed
chmod -R 755 emails

# 提供启动命令
echo -e "\n${GREEN}启动服务的命令:${NC}"
echo "  ./mailer -config config.json"

# 检查转发配置
echo -e "\n${YELLOW}检查邮件转发配置...${NC}"
FORWARD_ENABLED=$(grep -o '"forwardSMTP": *true' config.json)
if [ -z "$FORWARD_ENABLED" ]; then
    echo -e "${YELLOW}警告: 转发功能未启用，邮件将只保存在本地${NC}"
else
    FORWARD_HOST=$(grep -o '"forwardHost": *"[^"]*"' config.json | cut -d'"' -f4)
    if [ -z "$FORWARD_HOST" ]; then
        echo -e "${RED}错误: 转发功能已启用但未设置转发主机${NC}"
    else
        echo -e "${GREEN}转发主机设置为: $FORWARD_HOST${NC}"
        
        # 尝试解析域名
        if ! ping -c 1 $FORWARD_HOST >/dev/null 2>&1; then
            echo -e "${YELLOW}警告: 无法连接到转发主机 $FORWARD_HOST${NC}"
        fi
    fi
fi

echo -e "\n${GREEN}检查完成!${NC}"
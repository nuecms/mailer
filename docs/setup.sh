#!/bin/bash

# 创建文档目录结构
mkdir -p docs/.vitepress
mkdir -p docs/guides
mkdir -p docs/api
mkdir -p docs/examples
mkdir -p docs/public/images

# 将现有文档移动到新的目录结构
mv cloudflare_tunnel_setup.md docs/guides/cloudflare-tunnel.md
mv deploy_guide.md docs/guides/deployment.md
mv dkim_setup.md docs/guides/dkim-setup.md
mv advanced_features.md docs/guides/advanced-features.md
mv optimization_guide.md docs/guides/optimization.md

# 创建配置文件
touch docs/.vitepress/config.js
touch docs/index.md
touch docs/guides/configuration.md
touch docs/guides/troubleshooting.md
touch docs/examples/python.md
touch docs/examples/nodejs.md

# 设置正确的权限
chmod +x docs/setup.sh

echo "文档目录结构已创建"

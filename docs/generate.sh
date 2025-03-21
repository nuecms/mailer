#!/bin/bash

# 安装依赖
npm install

# 构建文档
npm run build

echo "文档构建完成！可以使用以下命令预览："
echo "cd docs && npm run preview"

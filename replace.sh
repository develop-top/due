#!/bin/bash
set -euxo pipefail  # 启用严格模式：报错停止、显示执行命令、检查未定义变量

# github.com/dobyte替换成github.com/develop-top
find . \( -name '*.go' -o -name '*.mod' \) -type f -exec sed -i 's/github\.com\/dobyte/github\.com\/develop-top/g' {} +

# 递归查找所有 go.mod 文件，逐个处理
find . -name "go.mod" -print0 | while IFS= read -r -d '' modfile; do
    (
        cd "$(dirname "$modfile")"  # 进入目录（处理空格）
        echo "→ 处理模块: $(pwd)"
        go mod tidy -v  # 显示详细日志
    ) || exit 1  # 任一子模块失败则终止脚本
done
#!/bin/bash
#start-ganache-and-save-keyssh

# 调试信息输出
set -x

# 启动 Ganache CLI 并将输出重定向到临时文件
ganache --mnemonic "dttp" > ganache_output.txt &

# 等待 Ganache CLI 完全启动
sleep 5

# 列出当前的 Node.js 进程
echo "Before killing process:"
pgrep -fl node

# 获取占用端口 8545 的进程 ID
GANACHE_PID=$(lsof -t -i:8545)
if [ ! -z "$GANACHE_PID" ]; then
  echo "Killing process $GANACHE_PID occupying port 8545"
  sudo kill $GANACHE_PID
  sleep 2
else
  echo "No process found occupying port 8545"
fi

# 检查进程是否成功终止
if sudo kill -0 $GANACHE_PID 2>/dev/null; then
  echo "Failed to kill process $GANACHE_PID"
else
  echo "Process $GANACHE_PID killed successfully"
fi

# 创建或清空现有的 .env 文件
echo "" > .env

# 读取私钥并写入到 .env 文件，去掉 '0x' 前缀
cat ganache_output.txt | grep 'Private Keys' -A 12 | grep -o '0x.*' | while read -r line; do
  echo "PRIVATE_KEY_$((++i))=${line:2}" >> .env
done

# 清理临时文件
rm ganache_output.txt

echo "Private keys saved to .env file."

#!/bin/bash
#start-ganache-and-save-keyssh

# 调试信息输出
set -x

# 启动 Ganache CLI 并将输出重定向到临时文件
OS=$(uname -s)
 
case "$OS" in
  Linux*)
    echo "Linux"
    ganache --mnemonic "dttp"  > ganache_output.txt &
    ;;
  Darwin*)
    echo "macOS"
    ganache-cli --mnemonic "dttp" > ganache_output.txt &
    ;;
  CYGWIN*|MINGW32*|MSYS*|MINGW*)
    echo "Windows"
    ;;
  *)
    echo "Unknown OS"
    ;;
esac

# 等待 Ganache CLI 完全启动
sleep 5
rm .env

# 提取可用账户并写入到 .env 文件
i=1
cat ganache_output.txt | grep -A 12 'Available Accounts' | grep '0x' | while read -r line; do
  address=$(echo $line | awk '{print $2}')
  echo "ACCOUNT_$i=$address" >> .env
  ((i++))
done
a=0
# 读取私钥并写入到 .env 文件，去掉 '0x' 前缀
cat ganache_output.txt | grep 'Private Keys' -A 12 | grep -o '0x.*' | while read -r line; do
  echo "PRIVATE_KEY_$((++a))=${line:2}" >> .env
done
#这个命令在ubuntu系统中存在不会杀死ganache的bug，ganache在ubuntu启动的进程为node，使用之后，需要手动kill杀死占用端口的进程
rm ganache_output.txt
#ubuntu可以使用这个命令：ps -ef | grep 'ganache' | grep -v grep | awk '{print $2}' | xargs kill -9
ps -ef|grep 'ganache'|xargs kill -9

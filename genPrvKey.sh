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

# 读取私钥并写入到 .env 文件，去掉 '0x' 前缀
cat ganache_output.txt | grep 'Private Keys' -A 12 | grep -o '0x.*' | while read -r line; do
  echo "PRIVATE_KEY_$((++i))=${line:2}" >> .env
done

rm ganache_output.txt
ps -ef|grep 'ganache'|xargs kill -9

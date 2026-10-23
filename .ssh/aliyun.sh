#!/bin/bash
# 连接阿里云 182.92.210.198 的辅助脚本
# 用法: ./aliyun.sh "远程命令"   或   ./aliyun.sh cmd  (从stdin读取)
# pscp 传输: pscp -batch -hostkey SHA256:... -pw PASS 本地文件 root@ip:路径
HOST="182.92.210.198"
USER="root"
PASS='{9$R!Q7J#B+2{D-&EM&S=g~4q>'
HOSTKEY="SHA256:VS/fv65W44WYQFdqFCQ4KNH9U5DKCjVseEz5kdcuwJg"

if [ -z "$1" ]; then
  echo "用法: $0 \"远程命令\""
  exit 1
fi

plink -ssh -batch -hostkey "$HOSTKEY" -l "$USER" -pw "$PASS" "$HOST" "$@"

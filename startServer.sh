#!/bin/bash

# 脚本的第一个参数用于指定配置文件的路径

# 用于在 shell 脚本退出时，删掉临时文件，结束子进程
trap "rm server;kill 0" EXIT

go build -o server
./server -cfg="$1" &

sleep 1
echo ">>> MyCache started successfully!"
wait
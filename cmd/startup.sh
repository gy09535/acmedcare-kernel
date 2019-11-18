#!/bin/bash
if [ ! -d "/go/logs" ]; then
  mkdir /go/logs
fi
if [ ! -f "/go/logs/start.log" ]; then
  touch "/go/logs/start.log"
fi
nohup go run /go/src/acmed.com/kernel/cmd/kernel.go>/go/logs/start.log 2>&1 &
echo "kernel is running,you can check /go/start.log"
sleep 10s
tail -f ${ServiceRoot}/logs/start.log
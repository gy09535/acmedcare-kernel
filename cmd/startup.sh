#!/bin/bash
if [ ! -d "/go/logs" ]; then
  mkdir /go/logs
fi
if [ ! -f "/go/logs/start.log" ]; then
  touch "/go/logs/start.log"
fi
nohup ./kernel>/go/logs/start.log 2>&1 &
echo "kernel is running,you can check /go/logs/start.log"
sleep 10s
tail -f ${ServiceRoot}/logs/start.log
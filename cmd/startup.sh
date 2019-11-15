#!/bin/bash
nohup go run /go/src/acmed.com/kernel/cmd/kernel.go>/go/logs/start.log 2>&1 &
echo "kernel is running,you can check /go/start.log"
tail -f ${ServiceRoot}/logs/start.log
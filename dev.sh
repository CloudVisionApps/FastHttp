#!/bin/bash

# Get all processes that are on port 80 and 443
PROCESS=$(lsof -i :80 | grep LISTEN)
PROCESS2=$(lsof -i :443 | grep LISTEN)

# If there is a process on port 80, kill it
if [ -n "$PROCESS" ]; then
  kill -9 $(lsof -t -i:80)
fi

# If there is a process on port 443, kill it
if [ -n "$PROCESS2" ]; then
  kill -9 $(lsof -t -i:443)
fi

go run fasthttp

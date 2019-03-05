#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Echo out all commands for monitoring progress
set -x

# Build all the things
go build -ldflags="-s -w" -o bin/config                     api/agent/config/main.go
go build -ldflags="-s -w" -o bin/hello                      api/agent/hello/main.go
go build -ldflags="-s -w" -o bin/admin                      api/admin
go build -ldflags="-s -w" -o bin/speedtestnetserverupdate   cron/speedtestnetserverupdate/main.go
go build -ldflags="-s -w" -o bin/alerts                     cron/alerts/main.go
go build -ldflags="-s -w" -o bin/dailysnapshot              cron/dailysnapshot/main.go
go build -ldflags="-s -w" -o bin/migrations                 cron/migrations/main.go
go build -ldflags="-s -w" -o bin/tasklog                    api/agent/tasklog/main.go


#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Echo out all commands for monitoring progress
set -x

# Build all the things
go build -ldflags="-s -w" -o bin/config                     api/agent/config/main.go
go build -ldflags="-s -w" -o bin/hello                      api/agent/hello/main.go
go build -ldflags="-s -w" -o bin/tag                        api/admin/tag/main.go
go build -ldflags="-s -w" -o bin/namedserver                api/admin/namedserver/main.go
go build -ldflags="-s -w" -o bin/node                       api/admin/node/main.go
go build -ldflags="-s -w" -o bin/report                     api/admin/report/main.go
go build -ldflags="-s -w" -o bin/reportingevent             api/admin/reportingevent/main.go
go build -ldflags="-s -w" -o bin/speedtestnetserver         api/admin/speedtestnetserver/main.go
go build -ldflags="-s -w" -o bin/speedtestnetserverupdate   cron/speedtestnetserverupdate/main.go
go build -ldflags="-s -w" -o bin/alerts                     cron/alerts/main.go
go build -ldflags="-s -w" -o bin/dailysnapshot              cron/dailysnapshot/main.go
go build -ldflags="-s -w" -o bin/migrations                 cron/migrations/main.go
go build -ldflags="-s -w" -o bin/tasklog                    api/agent/tasklog/main.go
go build -ldflags="-s -w" -o bin/user                       api/admin/user/main.go
go build -ldflags="-s -w" -o bin/version                    api/admin/version/main.go


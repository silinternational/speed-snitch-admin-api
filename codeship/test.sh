#!/usr/bin/env bash

# Exit script with error if any step fails.
#set -e
go get github.com/fillup/semver

sleep 10
AWS_ACCESS_KEY_ID=0 AWS_SECRET_ACCESS_KEY=0 go test ./...
# Because DynamoDB local is flaky, run again to be sure
AWS_ACCESS_KEY_ID=0 AWS_SECRET_ACCESS_KEY=0 go test ./...

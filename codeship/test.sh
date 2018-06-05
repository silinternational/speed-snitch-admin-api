#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e
go get github.com/fillup/semver

go test ./...

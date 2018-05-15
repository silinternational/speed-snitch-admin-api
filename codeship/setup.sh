#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

npm install -g serverless
npm install

# Install newer Go
export GO_VERSION=1.10.2
source /dev/stdin <<< "$(curl -sSL https://raw.githubusercontent.com/codeship/scripts/master/languages/go.sh)"

go get -u github.com/golang/dep/cmd/dep
dep ensure
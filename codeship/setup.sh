#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

npm install -g serverless
npm install

go get -u github.com/golang/dep/cmd/dep
dep ensure
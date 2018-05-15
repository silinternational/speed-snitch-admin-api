#!/usr/bin/env bash

# Exit script with error if any step fails.
set -x

npm install -g serverless
if [[ $? != 0 ]]; then
  exit $?;
fi

npm install
if [[ $? != 0 ]]; then
  exit $?;
fi

# Install newer Go
export GO_VERSION=1.10.2
wget https://raw.githubusercontent.com/codeship/scripts/master/languages/go.sh
chmod +x go.sh
source go.sh

go get -u github.com/golang/dep/cmd/dep
dep ensure
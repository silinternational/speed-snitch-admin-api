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
# strip all components from PATH which point to a GO installation and configure the download location
CLEANED_PATH=$(echo "${PATH}" | sed -r 's|/(usr/local\|tmp)/go(/([0-9]\.)+[0-9])?/bin:||g')
CACHED_DOWNLOAD="${HOME}/cache/go${GO_VERSION}.linux-amd64.tar.gz"

# configure the new GOROOT and PATH
export GOROOT="/tmp/go/${GO_VERSION}"
export PATH="${GOROOT}/bin:${CLEANED_PATH}"
mkdir -p "${GOROOT}"
wget --continue --output-document "${CACHED_DOWNLOAD}" "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz"
tar -xaf "${CACHED_DOWNLOAD}" --strip-components=1 --directory "${GOROOT}"

go get -u github.com/golang/dep/cmd/dep
dep ensure
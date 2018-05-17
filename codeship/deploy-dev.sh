#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/build.sh

# Export env vars
export CUSTOM_DOMAIN_NAME="${DEV_DOMAIN_NAME}"
export CERT_NAME="${DEV_CERT_NAME}"

cd api/admin
serverless deploy -v --stage dev

cd ../agent
serverless deploy -v --stage dev
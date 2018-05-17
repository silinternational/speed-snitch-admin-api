#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
./build.sh

# Export env vars
export CUSTOM_DOMAIN_NAME="${PROD_DOMAIN_NAME}"
export CERT_NAME="${PROD_CERT_NAME}"

cd api/admin
serverless deploy -v --stage prod

cd ../agent
serverless deploy -v --stage prod
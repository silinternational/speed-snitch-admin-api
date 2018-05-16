#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
./build.sh

cd api/admin
serverless deploy -v --stage dev

cd ../agent
serverless deploy -v --stage dev
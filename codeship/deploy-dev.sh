#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/build.sh

cd api/admin
serverless deploy -v --stage dev

cd ../agent
serverless deploy -v --stage dev
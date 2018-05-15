#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

cd api/admin
serverless deploy -v --stage dev

cd ../agent
serverless deploy -v --stage dev
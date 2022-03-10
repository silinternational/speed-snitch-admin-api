#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/build.sh

# Export env vars
export AGENT_API_TOKEN="${DEV_AGENT_API_TOKEN}"
export CUSTOM_DOMAIN_NAME="${DEV_DOMAIN_NAME}"
export CERT_NAME="${DEV_CERT_NAME}"
export DOWNLOAD_BASE_URL="${DEV_DOWNLOAD_BASE_URL}"
echo "DOWNLOAD_BASE_URL ... ${DOWNLOAD_BASE_URL} <<<"
export MYSQL_HOST="${DEV_MYSQL_HOST}"
export MYSQL_USER="${DEV_MYSQL_USER}"
export MYSQL_PASS="${DEV_MYSQL_PASS}"
export MYSQL_DB="${DEV_MYSQL_DB}"
export VPC_SG_ID="${DEV_VPC_SG_ID}"
export VPC_SUBNET1="${DEV_VPC_SUBNET1}"
export VPC_SUBNET2="${DEV_VPC_SUBNET2}"
export VPC_SUBNET3="${DEV_VPC_SUBNET3}"
export SES_RETURN_TO_ADDR="${DEV_SES_RETURN_TO_ADDR}"
export SES_AWS_REGION="${DEV_SES_AWS_REGION}"

# Echo commands to console
set -x

# Print the Serverless version in the logs
serverless --version

echo "Deploying stage dev..."

cd api/admin
serverless deploy --verbose --stage dev

cd ../agent
serverless deploy --verbose --stage dev

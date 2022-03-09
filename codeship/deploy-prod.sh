#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/build.sh

# Export env vars
export CUSTOM_DOMAIN_NAME="${PROD_DOMAIN_NAME}"
export CERT_NAME="${PROD_CERT_NAME}"
export DOWNLOAD_BASE_URL="${PROD_DOWNLOAD_BASE_URL}"
echo "DOWNLOAD_BASE_URL ... ${DOWNLOAD_BASE_URL} <<<"
export MYSQL_HOST="${PROD_MYSQL_HOST}"
export MYSQL_USER="${PROD_MYSQL_USER}"
export MYSQL_PASS="${PROD_MYSQL_PASS}"
export MYSQL_DB="${PROD_MYSQL_DB}"
export VPC_SG_ID="${PROD_VPC_SG_ID}"
export VPC_SUBNET1="${PROD_VPC_SUBNET1}"
export VPC_SUBNET2="${PROD_VPC_SUBNET2}"
export VPC_SUBNET3="${PROD_VPC_SUBNET3}"
export SES_RETURN_TO_ADDR="${PROD_SES_RETURN_TO_ADDR}"
export SES_AWS_REGION="${PROD_SES_AWS_REGION}"

# Echo commands to console
set -x

# Print the Serverless version in the logs
serverless --version

echo "Deploying stage prod..."

cd api/admin
#serverless deploy --verbose --stage prod

cd ../agent
serverless deploy --verbose --stage prod

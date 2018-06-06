#!/usr/bin/env ash

# Create data table
aws dynamodb create-table --table-name dataTable --attribute-definitions AttributeName=ID,AttributeType=S --key-schema AttributeName=ID,KeyType=HASH --provisioned-throughput ReadCapacityUnits=50,WriteCapacityUnits=50 --endpoint-url http://dynamo:8000
rc=$?;
if [[ $rc != 0 ]]; then
    exit $rc;
fi

# Create taskLog table
aws dynamodb create-table --table-name taskLogTable --attribute-definitions AttributeName=ID,AttributeType=S AttributeName=Timestamp,AttributeType=N --key-schema AttributeName=ID,KeyType=HASH AttributeName=Timestamp,KeyType=RANGE --provisioned-throughput ReadCapacityUnits=50,WriteCapacityUnits=50 --endpoint-url http://dynamo:8000
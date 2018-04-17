package db

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"fmt"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

func GetItem(tableName, attrName, attrValue string, returnObj interface{}) error {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			attrName: {
				S: aws.String(attrValue),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return nil
	}
	if result.Item == nil {
		return nil
	}

	// The result.Item object returned has the underlying type
	// map[string]*AttributeValue. We can use the UnmarshalMap helper
	// to parse this straight into the fields of a struct. Note:
	// UnmarshalListOfMaps also exists if you are working with multiple
	// items.
	//item := new(returnType)
	err = dynamodbattribute.UnmarshalMap(result.Item, returnObj)
	if err != nil {
		return err
	}

	return nil
}

func PutItem(tableName string, item interface{}) error {
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		domain.ServerError(fmt.Errorf("failed to DynamoDB marshal Record, %v", err))
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(domain.GetDbTableName(tableName)),
		Item: av,
	}

	_, err = db.PutItem(input)
	return err
}


func DeleteItem(tableName string, attributes map[string]*dynamodb.AttributeValue) (bool, error) {

	// Prepare the input for the query.
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: attributes,
	}

	// Delete the item from DynamoDB. I
	_, err := db.DeleteItem(input)

	if err.Error() == dynamodb.ErrCodeReplicaNotFoundException {
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return true, nil
}
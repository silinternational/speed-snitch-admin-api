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

//func GetItem(tableName, attrName, attrValue string, returnObj *interface{}) error {
//	// Prepare the input for the query.
//	input := &dynamodb.GetItemInput{
//		TableName: aws.String(tableName),
//		Key: map[string]*dynamodb.AttributeValue{
//			attrName: {
//				S: aws.String(attrValue),
//			},
//		},
//	}
//
//	// Retrieve the item from DynamoDB. If no matching item is found
//	// return nil.
//	result, err := db.GetItem(input)
//	if err != nil {
//		return nil, err
//	}
//	if result.Item == nil {
//		return nil, nil
//	}
//
//	// The result.Item object returned has the underlying type
//	// map[string]*AttributeValue. We can use the UnmarshalMap helper
//	// to parse this straight into the fields of a struct. Note:
//	// UnmarshalListOfMaps also exists if you are working with multiple
//	// items.
//	//item := new(returnType)
//	err = dynamodbattribute.UnmarshalMap(result.Item, returnObj)
//	if err != nil {
//		return nil, err
//	}
//
//	return nil
//}

func GetNode(macAddr string) (*domain.Node, error) {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(domain.GetDbTableName(domain.NodeTable)),
		Key: map[string]*dynamodb.AttributeValue{
			"MacAddr": {
				S: aws.String(macAddr),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	// The result.Item object returned has the underlying type
	// map[string]*AttributeValue. We can use the UnmarshalMap helper
	// to parse this straight into the fields of a struct. Note:
	// UnmarshalListOfMaps also exists if you are working with multiple
	// items.
	node := new(domain.Node)
	err = dynamodbattribute.UnmarshalMap(result.Item, node)
	if err != nil {
		return nil, err
	}

	return node, nil
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
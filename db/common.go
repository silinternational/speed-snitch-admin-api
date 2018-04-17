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

func GetItem(tableAlias, attrName, attrValue string, itemObj interface{}) error {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(domain.GetDbTableName(tableAlias)),
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
		return err
	}
	if result.Item == nil {
		return nil
	}

	// The result.Item object returned has the underlying type
	// map[string]*AttributeValue. We can use the UnmarshalMap helper
	// to parse this straight into the fields of a struct. Note:
	// UnmarshalListOfMaps also exists if you are working with multiple
	// items.
	err = dynamodbattribute.UnmarshalMap(result.Item, itemObj)
	if err != nil {
		return err
	}

	return nil
}

func PutItem(tableAlias string, item interface{}) error {
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		domain.ServerError(fmt.Errorf("failed to DynamoDB marshal Record, %v", err))
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(domain.GetDbTableName(tableAlias)),
		Item: av,
	}

	_, err = db.PutItem(input)
	return err
}

func DeleteItem(tableAlias, attrName, attrValue string) (bool, error) {

	// Prepare the input for the query.
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(domain.GetDbTableName(tableAlias)),
		Key: map[string]*dynamodb.AttributeValue{
			attrName: {
				S: aws.String(attrValue),
			},
		},
	}

	// Delete the item from DynamoDB. I
	_, err := db.DeleteItem(input)

	if err != nil && err.Error() == dynamodb.ErrCodeReplicaNotFoundException {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func scanTable(tableName string) ([]map[string]*dynamodb.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName: &tableName,
	}

	var results []map[string]*dynamodb.AttributeValue
	err := db.ScanPages(input,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			results = append(results, page.Items...)
			return !lastPage
		})

	if err != nil {
		return results, err
	}

	return results, nil
}

func ListTags() ([]domain.Tag, error) {

	var list []domain.Tag

	items, err := scanTable(domain.GetDbTableName(domain.TagTable))
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.Tag
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.Tag{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

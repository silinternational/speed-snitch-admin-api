package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"testing"
)

func FlushTables(t *testing.T) {
	tables := []string{"dataTable", "taskLogTable"}
	db := GetDb()

	for _, tableName := range tables {
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
			t.Error(err)
			t.Fail()
		}

		for _, item := range results {

			var keyCriteria map[string]*dynamodb.AttributeValue
			if tableName == "taskLogTable" {
				keyCriteria = map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(*item["ID"].S),
					},
					"Timestamp": {
						N: aws.String(*item["Timestamp"].N),
					},
				}
			} else {
				keyCriteria = map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(*item["ID"].S),
					},
				}
			}

			deleteInput := &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key:       keyCriteria,
			}

			_, err := db.DeleteItem(deleteInput)
			if err != nil {
				t.Errorf("Unable to delete item ID: %s, from table %s. Error: %s", *item["ID"].S, tableName, err.Error())
				t.Fail()
			}
		}
	}
}

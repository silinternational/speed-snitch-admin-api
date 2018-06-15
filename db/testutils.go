package db

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/silinternational/speed-snitch-admin-api"
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

func LoadTagFixtures(fixtures []domain.Tag, t *testing.T) {
	for _, item := range fixtures {
		err := PutItem(domain.DataTable, &item)
		if err != nil {
			t.Errorf("Error loading Tag fixture: %v\n%s", item, err.Error())
			t.Fail()
			return
		}
	}
}

func LoadNamedServerFixtures(fixtures []domain.NamedServer, t *testing.T) {
	for _, item := range fixtures {
		err := PutItem(domain.DataTable, &item)
		if err != nil {
			t.Errorf("Error loading NamedServer fixture: %v\n%s", item, err.Error())
			t.Fail()
			return
		}
	}
}

func LoadNodeFixtures(fixtures []domain.Node, t *testing.T) {
	for _, item := range fixtures {
		err := PutItem(domain.DataTable, &item)
		if err != nil {
			t.Errorf("Error loading Node fixture: %v\n%s", item, err.Error())
			t.Fail()
			return
		}
	}
}

func LoadSTNetServerListFixtures(fixtures []domain.STNetServerList, t *testing.T) {
	for _, item := range fixtures {
		err := PutItem(domain.DataTable, &item)
		if err != nil {
			t.Errorf("Error loading STNetServerList fixture: %v\n%s", item, err.Error())
			t.Fail()
			return
		}
	}
}

func LoadUserFixtures(fixtures []domain.User, t *testing.T) {
	for _, item := range fixtures {
		err := PutItem(domain.DataTable, &item)
		if err != nil {
			t.Errorf("Error loading User fixture: %v\n%s", item, err.Error())
			t.Fail()
			return
		}
	}
}

func areTagsEqual(expected, results []domain.Tag, t *testing.T) bool {
	errMsg := fmt.Sprintf("Tag slices are not equal.\nExpected: %v\n But got: %v", expected, results)

	if len(expected) != len(results) {
		t.Errorf(errMsg)
		return false
	}

	for _, nextExpected := range expected {
		foundMatch := false
		for _, nextResults := range results {
			if nextExpected == nextResults {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			t.Errorf(errMsg)
			return false
		}
	}

	return true
}

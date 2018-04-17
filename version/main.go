package main

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"net/http"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/lambda"
)

var dynamo = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, versionSpecified := req.PathParameters["version"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteVersion(req)
	case "GET":
		if versionSpecified {
			return viewVersion(req)
		}
		return listVersions(req)
	case "POST":
		return updateVersion(req)
	case "PUT":
		return updateVersion(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: " + req.HTTPMethod)
	}
}


func deleteVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	number := req.QueryStringParameters["number"]

	if number == "" {
		return domain.ClientError(http.StatusBadRequest, "number param must be specified")
	}

	success, err := db.DeleteItem(domain.VersionTable, "Number", number)

	if err != nil {
		return domain.ServerError(err)
	}

	if ! success {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func viewVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	number := req.QueryStringParameters["number"]

	if number == "" {
		return domain.ClientError(http.StatusBadRequest, "Number param must be specified")
	}

	var version domain.Version
	err := db.GetItem(domain.VersionTable, "Number", number, &version)
	if err != nil {
		return domain.ServerError(err)
	}

	if version.Description == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	js, err := json.Marshal(version)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listVersions(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var versions []domain.Version

	items, err := db.ScanTable(domain.GetDbTableName(domain.VersionTable))
	if err != nil {
		return domain.ServerError(err)
	}

	for _, item := range items {
		var itemObj domain.Version
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return domain.ServerError(err)
		}
		versions = append(versions, itemObj)
	}

	js, err := json.Marshal(versions)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func updateVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var version domain.Version

	// Get the version struct from the request body
	err := json.Unmarshal([]byte(req.Body), &version)
	if err != nil {
		return domain.ServerError(err)
	}

	// Update the version in the database
	err = db.PutItem(domain.VersionTable, version)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated version as json
	js, err := json.Marshal(version)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func main() {
	lambda.Start(router)
}

package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"strings"
)


var dynamo = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

func deleteNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr := req.QueryStringParameters["macAddr"]
	if !domain.IsValidMACAddress(macAddr) {
		return domain.ClientError(http.StatusBadRequest, "Bad Mac Address: " + macAddr	)
	}

	macAddr = strings.ToLower(macAddr)

	tableName := "Node"
	attributes := map[string]*dynamodb.AttributeValue{
		"MacAddr": {
			S: aws.String(macAddr),
		},
	}

	foundItem, err := db.DeleteItem(tableName, attributes)

	if ! foundItem {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "",
		}, nil
	}

	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func showNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func updateNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr := req.QueryStringParameters["macAddr"]
	if !domain.IsValidMACAddress(macAddr) {
		return domain.ClientError(http.StatusBadRequest, "Bad Mac Address: " + macAddr	)
	}

	macAddr = strings.ToLower(macAddr)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "DELETE":
		return deleteNode(req)
	case "GET":
		return showNode(req)
	case "PUT":
		return updateNode(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: " + req.HTTPMethod)
	}
}

func main() {
	lambda.Start(router)
}
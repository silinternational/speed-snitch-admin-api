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
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)


var dynamo = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))



func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, nodeSpecified := req.PathParameters["macAddr"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteNode(req)
	case "GET":
		if nodeSpecified {
			return viewNode(req)
		}
		return listNodes(req)
	case "POST":
		return updateNode(req)
	case "PUT":
		return updateNode(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: " + req.HTTPMethod)
	}
}


func deleteNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.QueryStringParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	success, err := db.DeleteItem(domain.NodeTable, "MacAddr", macAddr)

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

func viewNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.QueryStringParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var node domain.Node
	err = db.GetItem(domain.NodeTable, "MacAddr", macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.Arch == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	js, err := json.Marshal(node)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listNodes(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var nodes []domain.Node

	items, err := db.ScanTable(domain.GetDbTableName(domain.NodeTable))
	if err != nil {
		return domain.ServerError(err)
	}

	for _, item := range items {
		var itemObj domain.Node
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return domain.ServerError(err)
		}
		nodes = append(nodes, itemObj)
	}

	js, err := json.Marshal(nodes)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func updateNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var node domain.Node

	// Get the node struct from the request body
	err := json.Unmarshal([]byte(req.Body), &node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Clean the MAC Address
	macAddr, err := domain.CleanMACAddress(node.MacAddr)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}
	node.MacAddr = macAddr

	// Update the node in the database
	err = db.PutItem(domain.NodeTable, node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated node as json
	js, err := json.Marshal(node)
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
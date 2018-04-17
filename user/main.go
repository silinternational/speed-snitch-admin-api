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
	_, userSpecified := req.PathParameters["user"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteUser(req)
	case "GET":
		if userSpecified {
			return viewUser(req)
		}
		return listUsers(req)
	case "POST":
		return updateUser(req)
	case "PUT":
		return updateUser(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: " + req.HTTPMethod)
	}
}


func deleteUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := req.QueryStringParameters["id"]

	if id == "" {
		return domain.ClientError(http.StatusBadRequest, "id param must be specified")
	}

	success, err := db.DeleteItem(domain.UserTable, "ID", id)

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

func viewUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := req.QueryStringParameters["id"]

	if id == "" {
		return domain.ClientError(http.StatusBadRequest, "id param must be specified")
	}

	var user domain.User
	err := db.GetItem(domain.UserTable, "ID", id, &user)
	if err != nil {
		return domain.ServerError(err)
	}

	if user.Name == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	js, err := json.Marshal(user)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listUsers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var users []domain.User

	items, err := db.ScanTable(domain.GetDbTableName(domain.UserTable))
	if err != nil {
		return domain.ServerError(err)
	}

	for _, item := range items {
		var itemObj domain.User
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return domain.ServerError(err)
		}
		users = append(users, itemObj)
	}

	js, err := json.Marshal(users)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func updateUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var user domain.User

	// Get the user struct from the request body
	err := json.Unmarshal([]byte(req.Body), &user)
	if err != nil {
		return domain.ServerError(err)
	}

	// Update the user in the database
	err = db.PutItem(domain.UserTable, user)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated user as json
	js, err := json.Marshal(user)
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

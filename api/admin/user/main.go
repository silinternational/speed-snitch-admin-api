package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

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
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func deleteUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := req.PathParameters["id"]

	if id == "" {
		return domain.ClientError(http.StatusBadRequest, "id param must be specified")
	}

	success, err := db.DeleteItem(domain.DataTable, "user", id)

	if err != nil {
		return domain.ServerError(err)
	}

	if !success {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "",
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func viewUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := req.PathParameters["id"]

	if id == "" {
		return domain.ClientError(http.StatusBadRequest, "id param must be specified")
	}

	var user domain.User
	err := db.GetItem(domain.DataTable, "user", id, &user)
	if err != nil {
		return domain.ServerError(err)
	}

	if user.Name == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
	}

	js, err := json.Marshal(user)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func listUsers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	users, err := db.ListUsers()
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(users)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func updateUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var user domain.User

	// Get the user struct from the request body
	err := json.Unmarshal([]byte(req.Body), &user)
	if err != nil {
		return domain.ServerError(err)
	}

	if user.UserID == "" {
		return domain.ClientError(http.StatusBadRequest, "UserID is required")
	}

	user.ID = "user-" + user.UserID

	// Update the user in the database
	err = db.PutItem(domain.DataTable, user)
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
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func main() {
	lambda.Start(router)
}

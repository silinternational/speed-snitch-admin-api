package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, userSpecified := req.PathParameters["user"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteUser(req)
	case "GET":
		if userSpecified {
			if strings.HasSuffix(req.Path, "/tag") {
				return listUserTags(req)
			}
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
	id := req.PathParameters["uid"]

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
	uid := req.PathParameters["uid"]

	if uid == "" {
		return domain.ClientError(http.StatusBadRequest, "uid param must be specified")
	}

	var user domain.User
	err := db.GetItem(domain.DataTable, "user", uid, &user)
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

func listUserTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid := req.PathParameters["uid"]

	if uid == "" {
		return domain.ClientError(http.StatusBadRequest, "uid param must be specified")
	}

	var user domain.User
	err := db.GetItem(domain.DataTable, "user", uid, &user)
	if err != nil {
		return domain.ServerError(err)
	}

	allTags, err := db.ListTags()
	if err != nil {
		return domain.ServerError(err)
	}

	var userTags []domain.Tag

	for _, tag := range allTags {
		inArray, _ := domain.InArray(tag.UID, user.TagUIDs)
		if inArray {
			userTags = append(userTags, tag)
		}
	}

	js, err := json.Marshal(userTags)
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
	var user domain.User

	// If {uid} was provided in request, get existing record to update
	if req.PathParameters["uid"] != "" {
		err := db.GetItem(domain.DataTable, "user", req.PathParameters["uid"], &user)
		if err != nil {
			return domain.ServerError(err)
		}
	}

	// If UID is not set generate a UID
	if user.UID == "" {
		user.UID = domain.GetRandString(4)
	}
	user.ID = "user" + "-" + user.UID

	// Get the user struct from the request body
	var updatedUser domain.User
	err := json.Unmarshal([]byte(req.Body), &updatedUser)
	if err != nil {
		return domain.ServerError(err)
	}

	if updatedUser.Email == "" {
		return domain.ClientError(http.StatusBadRequest, "Email is required")
	}

	if !isValidRole(updatedUser.Role) {
		return domain.ClientError(http.StatusBadRequest, "Invalid Role provided")
	}

	// Make sure tags are valid and user calling api is allowed to use them
	if !db.AreTagsValid(updatedUser.TagUIDs) {
		return domain.ClientError(http.StatusBadRequest, "One or more submitted tags are invalid")
	}
	// @todo do we need to check if user making api call can use the tags provided?

	// Make sure user does not already exist with different UID
	exists, err := userAlreadyExists(user.UID, user.Email)
	if err != nil {
		return domain.ServerError(err)
	}
	if exists {
		return domain.ClientError(http.StatusConflict, "A user with this email already exists")
	}

	// Update user attributes
	user.Email = updatedUser.Email
	user.Name = updatedUser.Name
	user.Role = updatedUser.Role
	user.TagUIDs = updatedUser.TagUIDs

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

func isValidRole(role string) bool {
	if role == domain.UserRoleSuperAdmin || role == domain.UserRoleAdmin {
		return true
	}

	return false
}

// userAlreadyExist returns true if a user with the same email but different UID already exists
func userAlreadyExists(uid, email string) (bool, error) {
	allUsers, err := db.ListUsers()
	if err != nil {
		return false, err
	}

	for _, user := range allUsers {
		if user.Email == email && user.UID != uid {
			return true, nil
		}
	}

	return false, nil
}

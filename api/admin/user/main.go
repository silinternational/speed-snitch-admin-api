package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

const SelfType = domain.DataTypeUser

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, userSpecified := req.PathParameters["id"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteUser(req)
	case "GET":
		if userSpecified {
			return viewUser(req)
		}
		if strings.HasSuffix(req.Path, "/me") {
			return viewMe(req)
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

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var user domain.User
	err := db.DeleteItem(&user, id)
	return domain.ReturnJsonOrError(user, err)
}

func viewMe(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := db.GetUserFromRequest(req)
	if err != nil {
		user = domain.User{}
	}
	return domain.ReturnJsonOrError(user, err)
}

func viewUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var user domain.User
	err := db.GetItem(&user, id)
	return domain.ReturnJsonOrError(user, err)
}

func listUserTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var user domain.User
	err := db.GetItem(&user, id)
	return domain.ReturnJsonOrError(user.Tags, err)
}

func listUsers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var users []domain.User
	err := db.ListItems(&users, "name asc")
	return domain.ReturnJsonOrError(users, err)
}

func updateUser(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var user domain.User

	// If {id} was provided in request, get existing record to update
	if req.PathParameters["id"] != "" {
		id := domain.GetResourceIDFromRequest(req)
		if id == 0 {
			return domain.ClientError(http.StatusBadRequest, "Invalid ID")
		}

		err := db.GetItem(&user, id)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusNotFound,
					Body:       "",
				}, nil
			}
			return domain.ServerError(err)
		}
	}

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
	if !db.AreTagsValid(updatedUser.Tags) {
		return domain.ClientError(http.StatusBadRequest, "One or more submitted tags are invalid")
	}

	// Update user attributes
	user.Email = updatedUser.Email
	user.Name = updatedUser.Name
	user.Role = updatedUser.Role

	replacements := []domain.AssociationReplacement{
		{
			Replacement:     updatedUser.Tags,
			AssociationName: "Tags",
		},
	}

	// Update the user in the database
	err = db.PutItemWithAssociations(&user, replacements)
	return domain.ReturnJsonOrError(user, err)
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

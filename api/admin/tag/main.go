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

const UniqueNameErrorMessage = "Cannot update a Tag with a Name that is already in use."

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, tagSpecified := req.PathParameters["id"]
	switch req.HTTPMethod {
	case "GET":
		if tagSpecified {
			return viewTag(req)
		}
		return listTags(req)
	case "POST":
		return updateTag(req)
	case "PUT":
		return updateTag(req)
	case "DELETE":
		return deleteTag(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

func viewTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var tag domain.Tag
	err := db.GetItem(&tag, id)
	return domain.ReturnJsonOrError(tag, err)
}

func listTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var tags []domain.Tag
	err := db.ListItems(&tags, "name asc")
	return domain.ReturnJsonOrError(tags, err)
}

func updateTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var tag domain.Tag

	// If ID is provided, load existing tag for updating, otherwise we'll create a new one
	if req.PathParameters["id"] != "" {
		id := domain.GetResourceIDFromRequest(req)
		if id == 0 {
			return domain.ClientError(http.StatusBadRequest, "Invalid ID")
		}

		err := db.GetItem(&tag, id)
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

	// Parse request body for updated attributes
	var updatedTag domain.Tag
	err := json.Unmarshal([]byte(req.Body), &updatedTag)
	if err != nil {
		return domain.ServerError(err)
	}

	if updatedTag.Name == "" || updatedTag.Description == "" {
		return domain.ClientError(http.StatusUnprocessableEntity, "Name and Description are required")
	}

	// Update tag record attributes for persistence
	tag.Name = updatedTag.Name
	tag.Description = updatedTag.Description

	err = db.PutItem(&tag)
	if err != nil && strings.Contains(err.Error(), db.UniqueFieldErrorCode) {
		return domain.ClientError(http.StatusConflict, UniqueNameErrorMessage)
	}
	return domain.ReturnJsonOrError(tag, err)
}

func deleteTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var tag domain.Tag
	err := db.DeleteItem(&tag, id)
	return domain.ReturnJsonOrError(tag, err)
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

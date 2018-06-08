package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

const SelfType = domain.DataTypeTag

func main() {
	lambda.Start(router)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, tagSpecified := req.PathParameters["uid"]
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

	var tag domain.Tag
	err := db.GetItem(domain.DataTable, SelfType, req.PathParameters["uid"], &tag)
	if err != nil {
		return domain.ServerError(err)
	}

	if tag.Name == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	js, err := json.Marshal(tag)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	tags, err := db.ListTags()
	if err != nil {
		return domain.ServerError(err)
	}

	jsBody, err := domain.GetJSONFromSlice(tags)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       jsBody,
	}, nil
}

func updateTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var tag domain.Tag

	// If {uid} was provided in request, get existing record to update
	if req.PathParameters["uid"] != "" {
		err := db.GetItem(domain.DataTable, SelfType, req.PathParameters["uid"], &tag)
		if err != nil {
			return domain.ServerError(err)
		}
	}

	// If UID is not set generate a UID
	if tag.UID == "" {
		tag.UID = domain.GetRandString(4)
		tag.ID = SelfType + "-" + tag.UID
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

	// Make sure tag does not already exist with different UID
	exists, err := tagAlreadyExists(tag.UID, updatedTag.Name)
	if err != nil {
		return domain.ServerError(err)
	}
	if exists {
		return domain.ClientError(http.StatusConflict, "A tag with this name already exists")
	}

	// Update tag record attributes for persistence
	tag.Name = updatedTag.Name
	tag.Description = updatedTag.Description

	err = db.PutItem(domain.DataTable, tag)
	if err != nil {
		tagJson, _ := json.Marshal(tag)
		return domain.ServerError(fmt.Errorf("%s", tagJson))
		return domain.ServerError(err)
	}

	js, err := json.Marshal(tag)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func deleteTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	UID := req.PathParameters["uid"]

	deleted, err := db.DeleteItem(domain.DataTable, SelfType, UID)
	if err != nil {
		return domain.ServerError(err)
	}

	if !deleted && err == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "",
	}, nil
}

// tagAlreadyExist returns true if a tag with the same name but different UID already exists
func tagAlreadyExists(uid, name string) (bool, error) {
	allTags, err := db.ListTags()
	if err != nil {
		return false, err
	}

	for _, tag := range allTags {
		if tag.Name == name && tag.UID != uid {
			return true, nil
		}
	}

	return false, nil
}

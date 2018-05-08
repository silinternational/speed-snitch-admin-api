package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

func main() {
	lambda.Start(router)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, tagSpecified := req.PathParameters["name"]
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
	err := db.GetItem(domain.DataTable, "tag", req.PathParameters["name"], &tag)
	if err != nil {
		return domain.ServerError(err)
	}

	if tag.Name == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
	}

	js, err := json.Marshal(tag)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
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

	js, err := json.Marshal(tags)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func updateTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var tag domain.Tag
	err := json.Unmarshal([]byte(req.Body), &tag)
	if err != nil {
		return domain.ServerError(err)
	}

	if tag.Name == "" || tag.Description == "" {
		return domain.ClientError(http.StatusUnprocessableEntity, "Name and Description are required")
	}

	tag.ID = "tag-" + tag.Name

	err = db.PutItem(domain.DataTable, tag)
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(tag)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func deleteTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	name := req.PathParameters["name"]

	deleted, err := db.DeleteItem(domain.DataTable, "tag", name)
	if err != nil {
		return domain.ServerError(err)
	}

	if !deleted && err == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "",
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"net/http"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"encoding/json"
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
	var tag domain.Tag
	err := db.GetItem(domain.TagTable, "name", req.PathParameters["name"], &tag)
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
	}, nil
}

func updateTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var tag domain.Tag
	err := json.Unmarshal([]byte(req.Body), &tag)
	if err != nil {
		return domain.ServerError(err)
	}

	if tag.Name == "" || tag.Description == "" {
		return domain.ClientError(http.StatusUnprocessableEntity, "Name and Description are required")
	}

	err = db.PutItem(domain.TagTable, tag)
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
	}, nil
}

func deleteTag(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	name := req.PathParameters["name"]

	deleted, err := db.DeleteItem(domain.TagTable, "name", name)
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

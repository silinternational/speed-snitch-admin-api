package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

const SelfType = domain.DataTypeVersion

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, versionSpecified := req.PathParameters["number"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteVersion(req)
	case "GET":
		if versionSpecified {
			return viewVersion(req)
		}
		return listVersions(req)
	case "POST":
		return updateVersion(req)
	case "PUT":
		return updateVersion(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func deleteVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	number := req.PathParameters["number"]

	if number == "" {
		return domain.ClientError(http.StatusBadRequest, "number param must be specified")
	}

	success, err := db.DeleteItem(domain.DataTable, SelfType, number)

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

func viewVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	number := req.PathParameters["number"]

	if number == "" {
		return domain.ClientError(http.StatusBadRequest, "Number param must be specified")
	}

	var version domain.Version
	err := db.GetItem(domain.DataTable, SelfType, number, &version)
	if err != nil {
		return domain.ServerError(err)
	}

	if version.Description == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
	}

	js, err := json.Marshal(version)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func listVersions(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	versions, err := db.ListVersions()
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(versions)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func updateVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var version domain.Version

	// If {number} was provided in request, get existing record to update
	if req.PathParameters["number"] != "" {
		err := db.GetItem(domain.DataTable, SelfType, req.PathParameters["number"], &version)
		if err != nil {
			return domain.ServerError(err)
		}
	}

	// Get the version struct from the request body
	var updatedVersion domain.Version
	err := json.Unmarshal([]byte(req.Body), &updatedVersion)
	if err != nil {
		return domain.ServerError(err)
	}

	if version.Number == "" {
		version.Number = updatedVersion.Number
	}

	version.Description = updatedVersion.Description

	version.ID = SelfType + "-" + version.Number
	// Update the version in the database
	err = db.PutItem(domain.DataTable, version)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated version as json
	js, err := json.Marshal(version)
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

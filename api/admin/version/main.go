package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
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
	if req.PathParameters["id"] == "" {
		return domain.ClientError(http.StatusBadRequest, "Missing ID in path")
	}

	id := domain.GetUintFromString(req.PathParameters["id"])
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID in path")
	}

	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var version domain.Version
	err := db.DeleteItem(&version, req.PathParameters["id"])
	return domain.ReturnJsonOrError(version, err)
}

func viewVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.PathParameters["id"] == "" {
		return domain.ClientError(http.StatusBadRequest, "Missing ID in path")
	}

	id := domain.GetUintFromString(req.PathParameters["id"])
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID in path")
	}

	var version domain.Version
	err := db.GetItem(&version, req.PathParameters["id"])
	return domain.ReturnJsonOrError(version, err)
}

func listVersions(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var versions []domain.Version
	err := db.ListItems(&versions, "name asc")
	return domain.ReturnJsonOrError(versions, err)
}

func updateVersion(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var version domain.Version

	// If ID is provided, load existing tag for updating, otherwise we'll create a new one
	if req.PathParameters["id"] != "" {
		id := domain.GetUintFromString(req.PathParameters["id"])
		if id == 0 {
			return domain.ClientError(http.StatusBadRequest, "Invalid ID in path")
		}

		err := db.GetItem(&version, req.PathParameters["id"])
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

	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	// Parse request body for updated attributes
	var updatedVersion domain.Version
	err := json.Unmarshal([]byte(req.Body), &updatedVersion)
	if err != nil {
		return domain.ServerError(err)
	}

	if updatedVersion.Number == "" || updatedVersion.Description == "" {
		return domain.ClientError(http.StatusUnprocessableEntity, "Name and Description are required")
	}

	// Update tag record attributes for persistence
	version.Number = updatedVersion.Number
	version.Description = updatedVersion.Description

	err = db.PutItem(&version)
	return domain.ReturnJsonOrError(version, err)
}

func main() {
	lambda.Start(router)
}

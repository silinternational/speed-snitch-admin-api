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

const SelfType = domain.DataTypeNamedServer

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, serverSpecified := req.PathParameters["id"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteServer(req)
	case "GET":
		if serverSpecified {
			return viewServer(req)
		}
		return listServers(req)
	case "POST":
		return updateServer(req)
	case "PUT":
		return updateServer(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func deleteServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var server domain.NamedServer
	err := db.DeleteItem(&server, id)
	return domain.ReturnJsonOrError(server, err)
}

func viewServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var server domain.NamedServer
	err := db.GetItem(&server, id)
	return domain.ReturnJsonOrError(server, err)
}

func listServers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var servers []domain.NamedServer
	err := db.ListItems(&servers, "name asc")
	return domain.ReturnJsonOrError(servers, err)
}

func updateServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var server domain.NamedServer

	// If ID is provided, load existing server for updating, otherwise we'll create a new one
	if req.PathParameters["id"] != "" {
		id := domain.GetResourceIDFromRequest(req)
		if id == 0 {
			return domain.ClientError(http.StatusBadRequest, "Invalid ID")
		}

		err := db.GetItem(&server, id)
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

	// Get the NamedServer struct from the request body
	var updatedServer domain.NamedServer
	err := json.Unmarshal([]byte(req.Body), &updatedServer)
	if err != nil {
		return domain.ServerError(err)
	}

	server.ServerType = updatedServer.ServerType
	server.ServerHost = updatedServer.ServerHost
	server.ServerCountry = updatedServer.ServerCountry
	server.Name = updatedServer.Name
	server.Description = updatedServer.Description
	server.Notes = updatedServer.Notes

	var stnServer domain.SpeedTestNetServer
	if updatedServer.SpeedTestNetServerID != 0 {
		err = db.GetItem(&stnServer, updatedServer.SpeedTestNetServerID)
		if err != nil {
			return domain.ServerError(err)
		}
	}

	replacement := []domain.AssociationReplacement{
		{
			Replacement:     stnServer,
			AssociationName: "SpeedTestNetServer",
		},
	}

	// Update the namedserver in the database
	err = db.PutItemWithAssociations(&server, replacement)
	return domain.ReturnJsonOrError(server, err)
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

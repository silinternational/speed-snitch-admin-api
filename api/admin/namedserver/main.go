package main

import (
	"encoding/json"
	"fmt"
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
	if err != nil {
		domain.ReturnJsonOrError(server, err)
	}

	if server.ServerType != domain.ServerTypeSpeedTestNet {
		return domain.ReturnJsonOrError(server, err)
	}

	var stnServer domain.SpeedTestNetServer
	err = db.GetItem(&stnServer, server.SpeedTestNetServerID)
	if err != nil {
		return domain.ReturnJsonOrError(server, err)
	}

	server.ServerHost = stnServer.Host
	server.ServerCountry = stnServer.Country
	server.ServerCountryCode = stnServer.CountryCode
	return domain.ReturnJsonOrError(server, nil)
}

func listServers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var servers []domain.NamedServer

	namedServerType, exists := req.QueryStringParameters["type"]

	if !exists {
		err := db.ListItems(&servers, "name asc")
		return domain.ReturnJsonOrError(servers, err)
	}

	if namedServerType != domain.ServerTypeSpeedTestNet && namedServerType != domain.ServerTypeCustom {
		err := fmt.Errorf(
			`Invalid "type" query param. Must be %s or %s`,
			domain.ServerTypeSpeedTestNet,
			domain.ServerTypeCustom,
		)
		return domain.ReturnJsonOrError(servers, err)
	}

	servers, err := db.ListNamedServersByType(namedServerType)
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
	server.ServerCountryCode = updatedServer.ServerCountryCode
	server.Name = updatedServer.Name
	server.Description = updatedServer.Description
	server.Notes = updatedServer.Notes

	var stnServer domain.SpeedTestNetServer
	if updatedServer.ServerType == domain.ServerTypeSpeedTestNet {
		if updatedServer.SpeedTestNetServerID == 0 {
			err := fmt.Errorf("For server of type %s, the SpeedTestNetServerID cannot be 0.", domain.ServerTypeSpeedTestNet)
			return domain.ServerError(err)
		}

		err = db.GetItem(&stnServer, updatedServer.SpeedTestNetServerID)
		if err != nil {
			err := fmt.Errorf(
				"Error retrieving SpeedTestNet Server with ID %d.\n%s",
				updatedServer.SpeedTestNetServerID,
				err.Error(),
			)
			return domain.ServerError(err)
		}
		server.ServerHost = stnServer.Host
		server.ServerCountry = stnServer.Country
		server.ServerCountryCode = stnServer.CountryCode
	}

	replacement := []domain.AssociationReplacements{
		{
			Replacements:    []domain.SpeedTestNetServer{stnServer},
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

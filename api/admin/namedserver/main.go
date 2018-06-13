package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

const SelfType = domain.DataTypeNamedServer

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, serverSpecified := req.PathParameters["uid"]
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

	uid := req.PathParameters["uid"]

	success, err := db.DeleteItem(domain.DataTable, SelfType, uid)

	if err != nil {
		return domain.ServerError(err)
	}

	if !success {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func viewServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	uid := req.PathParameters["uid"]

	server, err := db.GetNamedServer(uid)
	if err != nil {
		return domain.ServerError(err)
	}

	if server.Name == "" {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	js, err := json.Marshal(server)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listServers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	servers, err := db.ListNamedServers()
	if err != nil {
		return domain.ServerError(err)
	}

	jsBody, err := domain.GetJSONFromSlice(servers)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       jsBody,
	}, nil
}

func updateServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var server domain.NamedServer

	// If {uid} was provided in request, get existing record to update
	if req.PathParameters["uid"] != "" {
		var err error
		server, err = db.GetNamedServer(req.PathParameters["uid"])
		if err != nil {
			return domain.ServerError(err)
		}
	}

	// If UID is not set generate a UID
	if server.UID == "" {
		server.UID = domain.GetRandString(4)
		server.ID = SelfType + "-" + server.UID
	}

	// Get the NamedServer struct from the request body
	var updatedServer domain.NamedServer
	err := json.Unmarshal([]byte(req.Body), &updatedServer)
	if err != nil {
		return domain.ServerError(err)
	}

	server.ServerType = updatedServer.ServerType
	server.SpeedTestNetServerID = updatedServer.SpeedTestNetServerID
	server.ServerHost = updatedServer.ServerHost
	server.Name = updatedServer.Name
	server.Description = updatedServer.Description
	server.Country = updatedServer.Country
	server.Notes = updatedServer.Notes

	// Update the namedserver in the database
	err = db.PutItem(domain.DataTable, server)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated namedserver as json
	js, err := json.Marshal(server)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func main() {
	lambda.Start(router)
}

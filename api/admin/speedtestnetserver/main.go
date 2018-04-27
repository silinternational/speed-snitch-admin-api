package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

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
	id := req.QueryStringParameters["id"]

	success, err := db.DeleteItem(domain.DataTable, "speedtestnetserver", id)

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

func viewServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := req.QueryStringParameters["id"]

	var server domain.SpeedTestNetServer
	err := db.GetItem(domain.DataTable, "speedtestnetserver", id, &server)
	if err != nil {
		return domain.ServerError(err)
	}

	if server.URL == "" {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	js, err := json.Marshal(server)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func listServers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	servers, err := db.ListSpeedTestNetServers()
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(servers)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func updateServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var server domain.SpeedTestNetServer

	// Get the SpeedTestNetServer struct from the request body
	err := json.Unmarshal([]byte(req.Body), &server)
	if err != nil {
		return domain.ServerError(err)
	}
	server.ID = "server-" + server.ServerID

	// Update the speedtestnetserver in the database
	err = db.PutItem(domain.DataTable, server)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated speedtestnetserver as json
	js, err := json.Marshal(server)
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

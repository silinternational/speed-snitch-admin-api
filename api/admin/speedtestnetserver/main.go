package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

const SelfType = domain.DataTypeSpeedTestNetServer

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, serverSpecified := req.PathParameters["serverID"]
	switch req.HTTPMethod {
	case "GET":
		if serverSpecified {
			return viewServer(req)
		}
		if strings.HasSuffix(req.Path, "/countries") {
			return listCountries(req)
		}
		return listServers(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func viewServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	serverID := req.PathParameters["serverID"]

	var server domain.SpeedTestNetServer
	err := db.GetItem(domain.DataTable, SelfType, serverID, &server)
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
	}, nil
}

func listServers(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	servers, err := db.ListSpeedTestNetServers()
	if err != nil {
		return domain.ServerError(err)
	}

	// If a specific country code was provided, limit results to just that country
	var serverList []domain.SpeedTestNetServer
	countryCode := req.QueryStringParameters["country"]
	if countryCode != "" {
		for _, server := range servers {
			if server.CountryCode == countryCode {
				serverList = append(serverList, server)
			}
		}
	} else {
		serverList = servers
	}

	js, err := json.Marshal(serverList)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listCountries(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var countries []domain.Country
	allServers, err := db.ListSpeedTestNetServers()
	if err != nil {
		return domain.ServerError(err)
	}

	for _, server := range allServers {
		country := domain.Country{
			Code: server.CountryCode,
			Name: server.Country,
		}
		inArray, _ := domain.InArray(country, countries)
		if !inArray {
			countries = append(countries, country)
		}
	}

	js, err := json.Marshal(countries)
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

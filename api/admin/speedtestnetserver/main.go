package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

const SelfType = domain.DataTypeSTNetServerList

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, serverSpecified := req.PathParameters["ID"]
	_, countrySpecified := req.PathParameters["countryCode"]
	switch req.HTTPMethod {
	case "GET":
		if strings.HasSuffix(req.Path, "/country") {
			return listCountries(req)
		}
		if countrySpecified && serverSpecified {
			return viewServer(req)
		}
		if countrySpecified {
			return listServersInCountry(req)
		}
		return domain.ClientError(
			http.StatusUnprocessableEntity,
			`"/country/" is required - optionally followed by the country code param.`,
		)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

// viewServer requires an "ID" path param.
func viewServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var server domain.SpeedTestNetServer
	err := db.GetItem(&server, id)
	return domain.ReturnJsonOrError(server, err)
}

// listServersInCountry requires a "countryCode" path param and returns the servers that have that country code
func listServersInCountry(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	countryCode := req.PathParameters["countryCode"]

	gdb, err := db.GetDb()
	if err != nil {
		return domain.ServerError(err)
	}

	var servers domain.SpeedTestNetServer

	gdb.Set("gorm:auto_preload", true).
		Where("countrycode = ?", countryCode).
		Order("name asc").
		Find(&servers)

	if gdb.Error != nil {
		return domain.ServerError(gdb.Error)
	}

	return domain.ReturnJsonOrError(servers, err)
}

func listCountries(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var countries []domain.Country
	err := db.ListItems(&countries, "code asc")
	return domain.ReturnJsonOrError(countries, err)
}

func main() {
	lambda.Start(router)
}

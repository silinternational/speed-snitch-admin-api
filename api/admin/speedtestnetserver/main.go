package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"sort"
	"strings"
)

const SelfType = domain.DataTypeSTNetServerList

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, serverSpecified := req.PathParameters["serverID"]
	_, countrySpecified := req.PathParameters["countryCode"]
	switch req.HTTPMethod {
	case "GET":
		if serverSpecified {
			return viewServer(req)
		}
		if countrySpecified {
			return listServersInCountry(req)
		}
		if strings.HasSuffix(req.Path, "/country") {
			return listCountries(req)
		}
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

// viewServer requires a "countryCode" and a "serverID" path param.
//  It gets the server row for that country code and extracts the server that
//  matches that serverID.
func viewServer(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	countryCode := req.PathParameters["countryCode"]
	var serversInCountry domain.STNetServerList
	err := db.GetItem(domain.DataTable, SelfType, countryCode, &serversInCountry)
	if err != nil {
		return domain.ServerError(err)
	}

	serverID := req.PathParameters["serverID"]

	var server domain.SpeedTestNetServer

	for _, countryServer := range serversInCountry.Servers {
		if countryServer.ServerID == serverID {
			server = countryServer
			break
		}
	}

	if server.Host == "" {
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

// listServersInCountry requires a "countryCode" path param and returns the one row that has the
//   list of servers for that country
func listServersInCountry(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	countryCode := req.PathParameters["countryCode"]

	var serversInCountry domain.STNetServerList
	err := db.GetItem(domain.DataTable, SelfType, countryCode, &serversInCountry)
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(serversInCountry)
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

	allServerLists, err := db.ListSTNetServerLists()
	if err != nil {
		return domain.ServerError(err)
	}

	countryList, err := getCountriesFromSTNetServerLists(allServerLists)
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(countryList)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func getCountriesFromSTNetServerLists(lists []domain.STNetServerList) ([]domain.Country, error) {
	countries := map[string]string{}

	for _, serverList := range lists {
		countries[serverList.Country.Code] = serverList.Country.Name
	}

	var countryList []domain.Country

	for code, name := range countries {
		countryList = append(countryList, domain.Country{Code: code, Name: name})
	}

	// Sort ascending by country name
	sort.Slice(countryList, func(i, j int) bool {
		return countryList[i].Name < countryList[j].Name
	})

	return countryList, nil
}

func main() {
	lambda.Start(router)
}

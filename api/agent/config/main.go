package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

func getConfig(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var node domain.Node
	err = db.GetItem(domain.NodeTable, "MacAddr", macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.Arch == "" {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	if node.ConfiguredVersion == "" || node.ConfiguredVersion == "latest" {
		latestVersion, err := db.GetLatestVersion()
		if err != nil {
			return domain.ServerError(err)
		}
		node.ConfiguredVersion = latestVersion.Number

	}

	downloadUrl := domain.GetUrlForAgentVersion(node.ConfiguredVersion, node.OS, node.Arch)
	config := domain.NodeConfig{
		Version: struct {
			Number string
			URL    string
		}{
			Number: node.ConfiguredVersion,
			URL:    downloadUrl,
		},
		Tasks: node.Tasks,
	}

	js, err := json.Marshal(config)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func main() {
	lambda.Start(getConfig)
}

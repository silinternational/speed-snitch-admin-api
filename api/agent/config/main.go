package main

import (
	"encoding/json"
	"fmt"
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

	node, err := db.GetNode(macAddr)
	if err != nil {
		return domain.ServerError(err)
	}

	// If node was not found in db, return 204 No Content
	if node.Arch == "" {
		return domain.ClientError(http.StatusNoContent, "")
	}

	if node.ConfiguredVersion == "" || node.ConfiguredVersion == "latest" {
		latestVersion, err := db.GetLatestVersion()
		if err != nil {
			return domain.ServerError(err)
		}
		node.ConfiguredVersion = latestVersion.Number
	}

	downloadUrl := domain.GetUrlForAgentVersion(node.ConfiguredVersion, node.OS, node.Arch)

	newTasks, err := db.GetUpdatedTasks(node.Tasks)
	if err != nil {
		return domain.ServerError(fmt.Errorf("Error updating tasks for node %s\n%s", macAddr, err.Error()))
	}

	config := domain.NodeConfig{
		Version: struct {
			Number string
			URL    string
		}{
			Number: node.ConfiguredVersion,
			URL:    downloadUrl,
		},
		Tasks: newTasks,
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

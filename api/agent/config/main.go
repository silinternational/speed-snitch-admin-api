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

func getConfig(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	node := domain.Node{
		MacAddr: macAddr,
	}

	err = db.FindOne(&node)
	if err == gorm.ErrRecordNotFound {
		return domain.ClientError(http.StatusNoContent, "Could not find node for macAddr: "+node.MacAddr)
	} else if err != nil {
		return domain.ServerError(err)
	}

	if node.ConfiguredVersion.Number == "" || node.ConfiguredVersion.Number == "latest" {
		latestVersion, err := db.GetLatestVersion()
		if err != nil {
			return domain.ServerError(err)
		}
		node.ConfiguredVersion = latestVersion
	}

	downloadUrl := domain.GetUrlForAgentVersion(node.ConfiguredVersion.Number, node.OS, node.Arch)

	config := domain.NodeConfig{
		Version: struct {
			Number string
			URL    string
		}{
			Number: node.ConfiguredVersion.Number,
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

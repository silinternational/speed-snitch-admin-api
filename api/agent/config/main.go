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

	var newTasks []domain.Task

	for _, oldTask := range node.Tasks {
		// Only modify tasks that involve a NamedServer
		if oldTask.NamedServer.ID == "" {
			newTasks = append(newTasks, oldTask)
			continue
		}

		namedServer, err := db.GetNamedServer(oldTask.NamedServer.UID)
		if err != nil {
			return domain.ServerError(fmt.Errorf("Error getting NamedServer ... %s", err.Error()))
		}

		newTask := oldTask

		// If it's not a SpeedTestNetServer, add the server info
		if namedServer.ServerType != domain.ServerTypeSpeedTestNet {
			newTask.SpeedTestNetServerID = ""
			newTask.ServerHost = namedServer.ServerHost
		} else {
			// Use the NamedServer to get the SpeedTestNetServer's info
			var speedtestnetserver domain.SpeedTestNetServer
			speedtestnetserver, err := db.GetSpeedTestNetServerFromNamedServer(namedServer)
			if err != nil {
				return domain.ServerError(fmt.Errorf("Error getting SpeedTestNetServer from NamedServer ... %s", err.Error()))
			}

			newTask.SpeedTestNetServerID = speedtestnetserver.ServerID
			newTask.ServerHost = speedtestnetserver.Host
		}

		newTasks = append(newTasks, newTask)
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

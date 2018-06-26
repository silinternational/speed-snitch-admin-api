package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"os"
)

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr := req.PathParameters["macAddr"]
	entryType := req.PathParameters["entryType"]

	var taskLogEntry domain.TaskLogEntry
	err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}

	// Set required attributes that are not part of post body
	taskLogEntry.ID = entryType + "-" + macAddr
	taskLogEntry.MacAddr = macAddr
	taskLogEntry.ExpirationTime = taskLogEntry.Timestamp + 31557600 // expire one year after log entry was created

	// Enrich log entry with node metadata details
	node, err := db.GetNode(macAddr)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, "Invalid Node MacAddr")
	} else {
		taskLogEntry.NodeLocation = node.Location
		taskLogEntry.NodeCoordinates = node.Coordinates
		taskLogEntry.NodeNetwork = node.Network
		taskLogEntry.NodeIPAddress = node.IPAddress
		taskLogEntry.NodeRunningVersion = node.RunningVersion
	}

	// Enrich speed test log entries with SpeedTestNet server details
	if entryType == domain.TaskTypeSpeedTest {
		speedTestServer, err := db.GetSTNetServer(taskLogEntry.ServerCountry, taskLogEntry.ServerID)
		if err != nil {
			// Just log it and not error out for now
			fmt.Fprintf(
				os.Stdout,
				"\nUnable to enrich task log entry for node %s. Country: %s, ServerID: %s. Err: %s",
				macAddr, taskLogEntry.ServerCountry, taskLogEntry.ServerID, err.Error())
		} else {
			taskLogEntry.ServerCountry = speedTestServer.Country
			taskLogEntry.ServerCoordinates = fmt.Sprintf("%s,%s", speedTestServer.Lat, speedTestServer.Lon)
			taskLogEntry.ServerName = speedTestServer.Name
		}
	}

	err = db.PutItem(domain.TaskLogTable, taskLogEntry)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

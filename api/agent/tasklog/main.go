package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
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

	// Enrich log entry with SpeedTestNet server details
	//var speedTestServer domain.SpeedTestNetServer
	//err := db.GetItem(domain.DataTable, "speedtestnetserver", taskLogEntry.ServerID, &speedTestServer)
	//if err != nil {
	//	return domain.ClientError(http.StatusBadRequest, "Invalid SpeedTestNetServer ID")
	//} else {
	//	taskLogEntry.ServerCountry = speedTestServer.Country
	//	taskLogEntry.ServerCoordinates = fmt.Sprintf("%s,%s", speedTestServer.Lat, speedTestServer.Lon)
	//	taskLogEntry.ServerSponsor = speedTestServer.Sponsor
	//}

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

	err = db.PutItem(domain.TaskLogTable, taskLogEntry)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

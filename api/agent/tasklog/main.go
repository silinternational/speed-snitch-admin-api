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
	var taskLogEntries []domain.TaskLogEntry
	err := json.Unmarshal([]byte(req.Body), &taskLogEntries)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}

	for _, entry := range taskLogEntries {

		// Set required attributes that are not part of post body
		entry.ID = entryType + "-" + macAddr
		entry.MacAddr = macAddr
		entry.ExpirationTime = entry.Timestamp + 31557600 // expire one year after log entry was created

		// Enrich log entry with SpeedTestNet server details
		//var speedTestServer domain.SpeedTestNetServer
		//err := db.GetItem(domain.DataTable, "speedtestnetserver", entry.ServerID, &speedTestServer)
		//if err != nil {
		//	return domain.ClientError(http.StatusBadRequest, "Invalid SpeedTestNetServer ID")
		//} else {
		//	entry.ServerCountry = speedTestServer.Country
		//	entry.ServerCoordinates = fmt.Sprintf("%s,%s", speedTestServer.Lat, speedTestServer.Lon)
		//	entry.ServerSponsor = speedTestServer.Sponsor
		//}

		// Enrich log entry with node metadata details
		var node domain.Node
		err = db.GetItem(domain.DataTable, "node", macAddr, &node)
		if err != nil {
			return domain.ClientError(http.StatusBadRequest, "Invalid Node MacAddr")
		} else {
			entry.NodeLocation = node.Location
			entry.NodeCoordinates = node.Coordinates
			entry.NodeNetwork = node.Network
			entry.NodeIPAddress = node.IPAddress
			entry.NodeRunningVersion = node.RunningVersion
		}

		err = db.PutItem(domain.TaskLogTable, entry)
	}

	// Return a response with a 200 OK status and the JSON book record
	// as the body.
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

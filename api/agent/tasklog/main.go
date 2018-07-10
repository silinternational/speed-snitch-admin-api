package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"os"
)

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr := req.PathParameters["macAddr"]
	entryType := req.PathParameters["entryType"]

	cleanMac, err := domain.CleanMACAddress(macAddr)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}

	node, err := db.GetNodeByMacAddr(cleanMac)
	if gorm.IsRecordNotFoundError(err) {
		return domain.ClientError(http.StatusNotFound, err.Error())
	} else if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	switch entryType {
	case domain.TaskTypeSpeedTest:
		var taskLogEntry domain.TaskLogSpeedTest
		err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
		if err != nil {
			return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
		}
		taskLogEntry.NodeID = node.Model.ID
		taskLogEntry.NodeLocation = node.Location
		taskLogEntry.NodeCoordinates = node.Coordinates
		taskLogEntry.NodeNetwork = node.Network
		taskLogEntry.NodeIPAddress = node.IPAddress
		taskLogEntry.NodeRunningVersion = node.RunningVersion

		if taskLogEntry.ServerID != "" {
			id := domain.GetUintFromString(taskLogEntry.ServerID)
			if id != 0 {
				var namedServer domain.NamedServer
				err = db.GetItem(&namedServer, id)
				if err != nil {
					// Just log it and not error out for now
					fmt.Fprintf(
						os.Stdout,
						"\nUnable to enrich task log entry for node %s. Country: %s, ServerID: %s. Err: %s",
						macAddr, taskLogEntry.ServerCountry, taskLogEntry.ServerID, err.Error())
				} else {
					taskLogEntry.ServerCountry = namedServer.SpeedTestNetServer.CountryCode
					taskLogEntry.ServerCoordinates = fmt.Sprintf("%s,%s", namedServer.SpeedTestNetServer.Lat, namedServer.SpeedTestNetServer.Lon)
					taskLogEntry.ServerName = namedServer.SpeedTestNetServer.Name
				}
			}
		}

		err = db.PutItem(&taskLogEntry)
		if err != nil {
			return domain.ServerError(err)
		}

	case domain.TaskTypePing:
		var taskLogEntry domain.TaskLogPingTest
		err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
		if err != nil {
			return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
		}
		taskLogEntry.NodeID = node.Model.ID
		taskLogEntry.NodeLocation = node.Location
		taskLogEntry.NodeCoordinates = node.Coordinates
		taskLogEntry.NodeNetwork = node.Network
		taskLogEntry.NodeIPAddress = node.IPAddress
		taskLogEntry.NodeRunningVersionID = node.RunningVersionID

		err = db.PutItem(&taskLogEntry)
		if err != nil {
			return domain.ServerError(err)
		}

	case domain.LogTypeDowntime:
		var taskLogEntry domain.TaskLogNetworkDowntime
		err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
		if err != nil {
			return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
		}
		taskLogEntry.NodeID = node.Model.ID
		taskLogEntry.NodeNetwork = node.Network
		taskLogEntry.NodeIPAddress = node.IPAddress

		err = db.PutItem(&taskLogEntry)
		if err != nil {
			return domain.ServerError(err)
		}

	case domain.LogTypeRestart:
		var taskLogEntry domain.TaskLogRestart
		err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
		if err != nil {
			return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
		}
		taskLogEntry.NodeID = node.Model.ID

		err = db.PutItem(&taskLogEntry)
		if err != nil {
			return domain.ServerError(err)
		}

	case domain.LogTypeError:
		var taskLogEntry domain.TaskLogError
		err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
		if err != nil {
			return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
		}
		taskLogEntry.NodeID = node.Model.ID
		taskLogEntry.NodeLocation = node.Location
		taskLogEntry.NodeCoordinates = node.Coordinates
		taskLogEntry.NodeNetwork = node.Network
		taskLogEntry.NodeIPAddress = node.IPAddress
		taskLogEntry.NodeRunningVersionID = node.RunningVersionID

		if taskLogEntry.ServerID != "" {
			id := domain.GetUintFromString(taskLogEntry.ServerID)
			if id != 0 {
				var namedServer domain.NamedServer
				err = db.GetItem(&namedServer, id)
				if err != nil {
					// Just log it and not error out for now
					fmt.Fprintf(
						os.Stdout,
						"\nUnable to enrich task log entry for node %s. Country: %s, ServerID: %s. Err: %s",
						macAddr, taskLogEntry.ServerCountry, taskLogEntry.ServerID, err.Error())
				} else {
					taskLogEntry.ServerCountry = namedServer.SpeedTestNetServer.CountryCode
					taskLogEntry.ServerCoordinates = fmt.Sprintf("%s,%s", namedServer.SpeedTestNetServer.Lat, namedServer.SpeedTestNetServer.Lon)
					taskLogEntry.ServerName = namedServer.SpeedTestNetServer.Name
				}
			}
		}

		err = db.PutItem(&taskLogEntry)
		if err != nil {
			return domain.ServerError(err)
		}

	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

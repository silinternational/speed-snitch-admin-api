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

func putSpeedTest(req events.APIGatewayProxyRequest, node domain.Node, macAddr string) (events.APIGatewayProxyResponse, error) {
	var taskLogEntry domain.TaskLogSpeedTest
	err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}
	taskLogEntry.NodeID = node.ID
	taskLogEntry.NodeLocation = node.Location
	taskLogEntry.NodeCoordinates = node.Coordinates
	taskLogEntry.NodeNetwork = node.Network
	taskLogEntry.NodeIPAddress = node.IPAddress
	taskLogEntry.NodeRunningVersion = node.RunningVersion

	if taskLogEntry.NamedServerID != 0 {
		var namedServer domain.NamedServer
		err = db.GetItem(&namedServer, taskLogEntry.NamedServerID)
		if err != nil {
			// Just log it and not error out for now
			fmt.Fprintf(
				os.Stdout,
				"\nUnable to enrich task log entry for node %s. Country: %s, NamedServerID: %v. Err: %s",
				macAddr, taskLogEntry.ServerCountry, taskLogEntry.NamedServerID, err.Error())
		} else {
			taskLogEntry.ServerCountry = namedServer.ServerCountry
			taskLogEntry.ServerCoordinates = fmt.Sprintf("%s,%s", namedServer.SpeedTestNetServer.Lat, namedServer.SpeedTestNetServer.Lon)
			taskLogEntry.ServerName = namedServer.SpeedTestNetServer.Name
			taskLogEntry.ServerHost = namedServer.ServerHost
		}
	}

	err = db.PutItem(&taskLogEntry)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func putPingTest(req events.APIGatewayProxyRequest, node domain.Node, macAddr string) (events.APIGatewayProxyResponse, error) {
	var taskLogEntry domain.TaskLogPingTest
	err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}
	taskLogEntry.NodeID = node.ID
	taskLogEntry.NodeLocation = node.Location
	taskLogEntry.NodeCoordinates = node.Coordinates
	taskLogEntry.NodeNetwork = node.Network
	taskLogEntry.NodeIPAddress = node.IPAddress
	taskLogEntry.NodeRunningVersionID = node.RunningVersionID

	if taskLogEntry.NamedServerID != 0 {
		var namedServer domain.NamedServer
		err = db.GetItem(&namedServer, taskLogEntry.NamedServerID)
		if err != nil {
			// Just log it and not error out for now
			fmt.Fprintf(
				os.Stdout,
				"\nUnable to enrich task log entry for node %s. Country: %s, NamedServerID: %v. Err: %s",
				macAddr, taskLogEntry.ServerCountry, taskLogEntry.NamedServerID, err.Error())
		} else {
			taskLogEntry.ServerCountry = namedServer.ServerCountry
			taskLogEntry.ServerName = namedServer.Name
			taskLogEntry.ServerHost = namedServer.ServerHost
		}
	}

	err = db.PutItem(&taskLogEntry)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func putDowntime(req events.APIGatewayProxyRequest, node domain.Node) (events.APIGatewayProxyResponse, error) {

	var taskLogEntry domain.TaskLogNetworkDowntime
	err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}
	taskLogEntry.NodeID = node.ID
	taskLogEntry.NodeNetwork = node.Network
	taskLogEntry.NodeIPAddress = node.IPAddress

	err = db.PutItem(&taskLogEntry)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func putRestart(req events.APIGatewayProxyRequest, node domain.Node) (events.APIGatewayProxyResponse, error) {
	var taskLogEntry domain.TaskLogRestart
	err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}
	taskLogEntry.NodeID = node.ID

	err = db.PutItem(&taskLogEntry)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func putError(req events.APIGatewayProxyRequest, node domain.Node, macAddr string) (events.APIGatewayProxyResponse, error) {
	var taskLogEntry domain.TaskLogError
	err := json.Unmarshal([]byte(req.Body), &taskLogEntry)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}
	taskLogEntry.NodeID = node.ID
	taskLogEntry.NodeLocation = node.Location
	taskLogEntry.NodeCoordinates = node.Coordinates
	taskLogEntry.NodeNetwork = node.Network
	taskLogEntry.NodeIPAddress = node.IPAddress
	taskLogEntry.NodeRunningVersionID = node.RunningVersionID

	if taskLogEntry.NamedServerID != 0 {
		var namedServer domain.NamedServer
		err = db.GetItem(&namedServer, taskLogEntry.NamedServerID)
		if err != nil {
			// Just log it and not error out for now
			fmt.Fprintf(
				os.Stdout,
				"\nUnable to enrich task log entry for node %s. Country: %s, NamedServerID: %v. Err: %s",
				macAddr, taskLogEntry.ServerCountry, taskLogEntry.NamedServerID, err.Error())
		} else {
			taskLogEntry.ServerCountry = namedServer.ServerCountry
			taskLogEntry.ServerName = namedServer.Name
			taskLogEntry.ServerHost = namedServer.ServerHost
		}
	}

	err = db.PutItem(&taskLogEntry)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

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
		return putSpeedTest(req, node, macAddr)

	case domain.TaskTypePing:
		return putPingTest(req, node, macAddr)

	case domain.LogTypeDowntime:
		return putDowntime(req, node)

	case domain.LogTypeRestart:
		return putRestart(req, node)

	case domain.LogTypeError:
		return putError(req, node, macAddr)

	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	defer db.Db.Close()
	lambda.Start(Handler)
}

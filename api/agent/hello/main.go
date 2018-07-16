package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/ipinfo"
	"net/http"
	"time"
)

type Response struct {
	Message string `json:"message"`
}

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Check for existing agent or create new
	// Update attributes for agent
	// Save changes
	// Return 204

	// Parse request body
	var helloReq domain.HelloRequest
	err := json.Unmarshal([]byte(req.Body), &helloReq)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}

	node := domain.Node{
		MacAddr: helloReq.ID,
	}

	err = db.FindOne(&node)
	if err == gorm.ErrRecordNotFound {
		node.OS = helloReq.OS
		node.Arch = helloReq.Arch
		node.FirstSeen = domain.GetTimeNow()
	} else if err != nil {
		return domain.ServerError(err)
	}

	// If node is new or IP address has changed, update ip address, location, and coordinates
	var reqSourceIP string
	_, ok := req.Headers["CF-Connecting-IP"]
	if ok {
		reqSourceIP = req.Headers["CF-Connecting-IP"]
	} else {
		reqSourceIP = req.RequestContext.Identity.SourceIP
	}

	if node.IPAddress != reqSourceIP {
		node.IPAddress = reqSourceIP
		ipDetails, err := ipinfo.GetIPInfo(reqSourceIP)
		if err != err {
			return domain.ServerError(err)
		}
		node.Location = fmt.Sprintf("%s, %s, %s", ipDetails.Country, ipDetails.Region, ipDetails.City)
		node.Coordinates = ipDetails.Loc
	}

	version := domain.Version{
		Number: helloReq.Version,
	}

	err = db.FindOne(&version)
	if err == gorm.ErrRecordNotFound {
		errMsg := fmt.Sprintf("\nRunning Version not found: %s\n\t(for node %s)\n", helloReq.Version, helloReq.ID)
		domain.ErrorLogger.Println(errMsg)
	} else if err != nil {
		return domain.ServerError(err)
	} else {
		node.RunningVersion = version
	}

	// Update transient fields
	node.Uptime = helloReq.Uptime
	node.LastSeen = domain.GetTimeNow()

	// Write to DB
	err = db.PutItem(&node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return a response with a 204 status
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	defer db.Db.Close()
	lambda.Start(Handler)
}

func getTimeNow() int64 {
	utcNow := time.Now().UTC()
	return utcNow.Unix()
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

	// Fetch existing node if exists
	node, err := db.GetNode(helloReq.ID)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.MacAddr == "" {
		// Initialize new node record
		node.ID = "node-" + helloReq.ID
		node.MacAddr = helloReq.ID
		node.OS = helloReq.OS
		node.Arch = helloReq.Arch
		node.FirstSeen = getTimeNow()
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

	// Update transient fields
	node.RunningVersion = helloReq.Version
	node.Uptime = helloReq.Uptime
	node.LastSeen = getTimeNow()

	// Write to DB
	err = db.PutItem(domain.DataTable, node)
	if err != nil {
		return domain.ServerError(err)
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

func getTimeNow() string {
	t := time.Now().UTC()
	return t.Format(time.RFC3339)
}

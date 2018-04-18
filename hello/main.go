package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/ipinfo"
	"net/http"
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
	var node domain.Node
	err = db.GetItem(domain.NodeTable, "MacAddr", helloReq.ID, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.MacAddr == "" {
		// Initialize new node record
		node.MacAddr = helloReq.ID
		node.OS = helloReq.OS
		node.Arch = helloReq.Arch
	}

	// If node is new or IP address has changed, update ip address, location, and coordinates
	reqSourceIP := req.RequestContext.Identity.SourceIP
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

	// Write to DB
	err = db.PutItem(domain.NodeTable, node)
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

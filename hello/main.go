package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"encoding/json"
	"github.com/silinternational/speed-snitch-admin-api/db"
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
	//node, err := db.GetNode(helloReq.ID)
	var node domain.Node
	err = db.GetItem(domain.NodeTable, "MacAddr", helloReq.ID, node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.MacAddr == "" {
		// Initialize new node record
		node.MacAddr = helloReq.ID
		node.OS = helloReq.OS
		node.Arch = helloReq.Arch
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

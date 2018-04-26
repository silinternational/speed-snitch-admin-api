package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, nodeSpecified := req.PathParameters["macAddr"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteNode(req)
	case "GET":
		if nodeSpecified {
			return viewNode(req)
		}
		return listNodes(req)
	case "POST":
		return updateNode(req)
	case "PUT":
		return updateNode(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func deleteNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	success, err := db.DeleteItem(domain.NodeTable, "MacAddr", macAddr)

	if err != nil {
		return domain.ServerError(err)
	}

	if !success {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func viewNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var node domain.Node
	err = db.GetItem(domain.NodeTable, "MacAddr", macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.Arch == "" {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	// Ensure user is authorized ...
	httpStatus, err := isUserForbidden(req, node)
	if err != nil {
		return domain.ServerError(err)
	}

	if httpStatus > 0 {
		return domain.ClientError(httpStatus, http.StatusText(httpStatus))
	}

	js, err := json.Marshal(node)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listNodes(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	nodes, err := db.ListNodes()
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(nodes)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func updateNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var node domain.Node

	// Get the node struct from the request body
	err := json.Unmarshal([]byte(req.Body), &node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Clean the MAC Address
	macAddr, err := domain.CleanMACAddress(node.MacAddr)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}
	node.MacAddr = macAddr

	// Update the node in the database
	err = db.PutItem(domain.NodeTable, node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated node as json
	js, err := json.Marshal(node)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func main() {
	lambda.Start(router)
}

func isUserForbidden(req events.APIGatewayProxyRequest, node domain.Node) (int, error) {
	// Ensure user is authorized ...
	// Get the user's ID
	userID, ok := req.Headers[domain.UserReqHeaderID]
	if !ok {
		return http.StatusUnauthorized, nil
	}

	// Get the user
	var user domain.User
	err := db.GetItem(domain.UserTable, "ID", userID, &user)
	if err != nil {
		return 0, err
	}

	// Forbid the user if inadequate permissions
	if !domain.CanUserUseNode(user, node) {
		return http.StatusForbidden, nil
	}

	return 0, nil
}

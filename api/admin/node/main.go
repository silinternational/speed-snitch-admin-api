package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, nodeSpecified := req.PathParameters["macAddr"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteNode(req)
	case "GET":
		if nodeSpecified {
			if strings.HasSuffix(req.Path, "/tag") {
				return listNodeTags(req)
			}
			return viewNode(req)
		}
		return listNodes(req)
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

	success, err := db.DeleteItem(domain.DataTable, "node", macAddr)

	if err != nil {
		return domain.ServerError(err)
	}

	if !success {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "",
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func viewNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var node domain.Node
	err = db.GetItem(domain.DataTable, "node", macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.Arch == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
			Headers:    domain.DefaultResponseCorsHeaders,
		}, nil
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
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func listNodeTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var node domain.Node
	err = db.GetItem(domain.DataTable, "node", macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	allTags, err := db.ListTags()
	if err != nil {
		return domain.ServerError(err)
	}

	var nodeTags []domain.Tag

	for _, tag := range allTags {
		inArray, _ := domain.InArray(tag.UID, node.TagUIDs)
		if inArray {
			nodeTags = append(nodeTags, tag)
		}
	}

	js, err := json.Marshal(nodeTags)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
		Headers:    domain.DefaultResponseCorsHeaders,
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
		Headers:    domain.DefaultResponseCorsHeaders,
	}, nil
}

func updateNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var node domain.Node

	if req.PathParameters["macAddr"] == "" {
		return domain.ClientError(http.StatusBadRequest, "Mac Address is required")
	}

	// Clean the MAC Address
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}
	err = db.GetItem(domain.DataTable, "node", macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Get the node struct from the request body
	var updatedNode domain.Node
	err = json.Unmarshal([]byte(req.Body), &updatedNode)
	if err != nil {
		return domain.ServerError(err)
	}

	// Make sure tags are valid and user calling api is allowed to use them
	if !db.AreTagsValid(updatedNode.TagUIDs) {
		return domain.ClientError(http.StatusBadRequest, "One or more submitted tags are invalid")
	}
	// @todo do we need to check if user making api call can use the tags provided?

	// Apply updates to node
	node.ID = "node-" + macAddr
	node.ConfiguredVersion = updatedNode.ConfiguredVersion
	node.Tasks = updatedNode.Tasks
	node.Contacts = updatedNode.Contacts
	node.TagUIDs = updatedNode.TagUIDs
	node.Nickname = updatedNode.Nickname
	node.Notes = updatedNode.Notes

	// Update the node in the database
	err = db.PutItem(domain.DataTable, node)
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
		Headers:    domain.DefaultResponseCorsHeaders,
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
	err := db.GetItem(domain.DataTable, "user", userID, &user)
	if err != nil {
		return 0, err
	}

	// Forbid the user if inadequate permissions
	if !domain.CanUserUseNode(user, node) {
		return http.StatusForbidden, nil
	}

	return 0, nil
}

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
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

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
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
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

	user, err := db.GetUserFromRequest(req)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var js []byte

	if user.Role == domain.UserRoleSuperAdmin {
		js, err = json.Marshal(nodes)
		if err != nil {
			return domain.ServerError(err)
		}
	} else {
		visibleNodes := []domain.Node{}
		for _, node := range nodes {
			if domain.CanUserUseNode(user, node) {
				visibleNodes = append(visibleNodes, node)
			}
		}

		js, err = json.Marshal(visibleNodes)
		if err != nil {
			return domain.ServerError(err)
		}
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

	// If node already exists, ensure user is authorized ...
	var existingNode domain.Node
	err = db.GetItem(domain.DataTable, "node", macAddr, &existingNode)
	if err == nil {
		statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, existingNode.Tags)
		if statusCode > 0 {
			return domain.ClientError(statusCode, errMsg)
		}
	}

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

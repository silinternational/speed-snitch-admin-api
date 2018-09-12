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
	"strings"
)

const UniqueNameErrorMessage = "Cannot update a Reporting Event with a Name that is already in use."

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, eventSpecified := req.PathParameters["id"]
	switch req.HTTPMethod {
	case "GET":
		if eventSpecified {
			return viewEvent(req)
		}
		return listEvents(req)
	case "POST":
		return updateEvent(req)
	case "PUT":
		return updateEvent(req)
	case "DELETE":
		return deleteEvent(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}
}

func getAuthStatusForEvent(req events.APIGatewayProxyRequest, event domain.ReportingEvent) (int, string) {
	// Only SuperAdmins can deal with app level (nodeless) events
	if event.NodeID == 0 {
		return db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})

	}

	// Ensure user has a tag that matches this event's node's tags
	return db.GetAuthorizationStatus(req, domain.PermissionTagBased, event.Node.Tags)
}

func viewEvent(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var reportingEvent domain.ReportingEvent
	err := db.GetItem(&reportingEvent, id)

	// Enforce user authorization for the event
	statusCode, errMsg := getAuthStatusForEvent(req, reportingEvent)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	return domain.ReturnJsonOrError(reportingEvent, err)
}

func listEvents(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	nodeID := uint(0)
	nodeParam, exists := req.QueryStringParameters["node_id"]
	if exists {
		nodeID = domain.GetUintFromString(nodeParam)
	}

	if nodeID > 0 {
		return listEventsForNode(req, nodeID)
	}

	// Just return global events
	globalEvents, err := db.GetReportingEvents(0)
	if err != nil {
		err := fmt.Errorf("Error getting global reporting events. %s", err.Error())
		return domain.ReturnJsonOrError([]domain.ReportingEvent{}, err)
	}

	return domain.ReturnJsonOrError(globalEvents, err)
}

func listEventsForNode(req events.APIGatewayProxyRequest, nodeID uint) (events.APIGatewayProxyResponse, error) {
	user, err := db.GetUserFromRequest(req)
	if err != nil {
		errMsg := fmt.Sprintf("error getting user from database: %s", err.Error())
		return domain.ClientError(http.StatusBadRequest, errMsg)
	}

	node := domain.Node{}
	err = db.GetItem(&node, nodeID)
	if err != nil {
		errMsg := fmt.Sprintf("error getting node from database (id %v): %s", nodeID, err.Error())
		return domain.ClientError(http.StatusBadRequest, errMsg)
	}

	if !domain.CanUserUseNode(user, node) {
		return domain.ClientError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	eventsForNode, err := db.GetReportingEvents(nodeID)
	if err != nil {
		return domain.ReturnJsonOrError([]domain.ReportingEvent{}, err)
	}

	return domain.ReturnJsonOrError(eventsForNode, err)
}

func updateEvent(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var reportingEvent domain.ReportingEvent

	// If ID is provided, load existing ReportingEvent for updating, otherwise we'll create a new one
	if req.PathParameters["id"] != "" {
		id := domain.GetResourceIDFromRequest(req)
		if id == 0 {
			return domain.ClientError(http.StatusBadRequest, "Invalid ID")
		}

		err := db.GetItem(&reportingEvent, id)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusNotFound,
					Body:       "",
				}, nil
			}
			return domain.ServerError(err)
		}
	}

	// Parse request body for updated attributes
	var updatedEvent domain.ReportingEvent

	err := json.Unmarshal([]byte(req.Body), &updatedEvent)
	if err != nil {
		return domain.ServerError(err)
	}

	// Enforce user authorization for the new version of the event
	statusCode, errMsg := getAuthStatusForEvent(req, updatedEvent)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	// Enforce user authorization for the old version of the event
	statusCode, errMsg = getAuthStatusForEvent(req, reportingEvent)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	if updatedEvent.Name == "" || updatedEvent.Date == "" {
		return domain.ClientError(http.StatusUnprocessableEntity, "Name and Date are required")
	}

	reportingEvent.Date = updatedEvent.Date
	err = reportingEvent.SetTimestamp()
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	reportingEvent.Name = updatedEvent.Name
	reportingEvent.Description = updatedEvent.Description

	err = db.PutItem(&reportingEvent)
	if err != nil && strings.Contains(err.Error(), db.UniqueFieldErrorCode) {
		return domain.ClientError(http.StatusConflict, UniqueNameErrorMessage)
	}
	return domain.ReturnJsonOrError(reportingEvent, err)
}

func deleteEvent(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var reportingEvent domain.ReportingEvent

	// Enforce user authorization for the event
	statusCode, errMsg := getAuthStatusForEvent(req, reportingEvent)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	err := db.DeleteItem(&reportingEvent, id)
	return domain.ReturnJsonOrError(reportingEvent, err)
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

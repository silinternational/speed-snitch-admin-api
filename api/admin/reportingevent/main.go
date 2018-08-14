package main

import (
	"encoding/json"
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

func viewEvent(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var reportingEvent domain.ReportingEvent
	err := db.GetItem(&reportingEvent, id)
	return domain.ReturnJsonOrError(reportingEvent, err)
}

func listEvents(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	var reportingEvents []domain.ReportingEvent
	err := db.ListItems(&reportingEvents, "timestamp asc")
	return domain.ReturnJsonOrError(reportingEvents, err)
}

func updateEvent(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

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

	if updatedEvent.Name == "" {
		return domain.ClientError(http.StatusUnprocessableEntity, "Name is required")
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
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var reportingEvent domain.ReportingEvent
	err := db.DeleteItem(&reportingEvent, id)
	return domain.ReturnJsonOrError(reportingEvent, err)
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/reporting"
	"net/http"
	"time"
)

const PeriodTimeFormat = "2006-01-02"

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := req.PathParameters["id"]
	if id != "" {
		return viewNodeReport(req)
	}

	return domain.ClientError(http.StatusBadRequest, "id is required in url")
}

func viewNodeReport(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	// Validate Inputs
	interval := req.QueryStringParameters["interval"]
	if !reporting.IsValidReportingInterval(interval) {
		return domain.ClientError(http.StatusBadRequest, "Invalid interval specified")
	}

	periodStartTimestamp, err := getTimestampFromString(req.QueryStringParameters["start"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp, err := getTimestampFromString(req.QueryStringParameters["end"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	// Fetch node to ensure exists and get tags for authorization
	var node domain.Node
	err = db.GetItem(&node, id)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Ensure user is authorized ...
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	// Fetch snapshots
	snapshots, err := db.GetSnapshotsForRange(interval, id, periodStartTimestamp, periodEndTimestamp)
	return domain.ReturnJsonOrError(snapshots, err)
}

func getTimestampFromString(date string) (int64, error) {
	if date == "" {
		return 0, fmt.Errorf("parameters start and end are required and should be format YYYY-MM-DD")
	}
	dateTime, err := time.Parse(PeriodTimeFormat, date)
	if err != nil {
		return 0, err
	}
	timestamp, _, err := reporting.GetStartEndTimestampsForDate(dateTime)
	if err != nil {
		return 0, err
	}

	return timestamp, nil
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

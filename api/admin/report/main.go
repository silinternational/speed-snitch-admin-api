package main

import (
	"encoding/json"
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
	macAddr := req.PathParameters["macAddr"]
	if macAddr != "" {
		return viewNodeReport(req)
	}

	return domain.ClientError(http.StatusBadRequest, "macAddr is required in url")
}

func viewNodeReport(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr := req.PathParameters["macAddr"]

	// Validate Inputs
	interval := req.QueryStringParameters["interval"]
	if !reporting.IsValidReportingInterval(interval) {
		return domain.ClientError(http.StatusBadRequest, "Invalid interval specified")
	}

	periodStart := req.QueryStringParameters["start"]
	if periodStart == "" {
		return domain.ClientError(http.StatusBadRequest, "Parameter start is required and should be format YYYY-MM-DD")
	}
	periodStartTime, err := time.Parse(PeriodTimeFormat, periodStart)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}
	periodStartTimestamp, _, err := reporting.GetStartEndTimestampsForDate(periodStartTime)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEnd := req.QueryStringParameters["end"]
	if periodEnd == "" {
		return domain.ClientError(http.StatusBadRequest, "Parameter end is required and should be format YYYY-MM-DD")
	}
	periodEndTime, err := time.Parse(PeriodTimeFormat, periodEnd)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}
	_, periodEndTimestamp, err := reporting.GetStartEndTimestampsForDate(periodEndTime)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	// Fetch node to ensure exists and get tags for authorization
	var node domain.Node
	err = db.GetItem(domain.DataTable, domain.DataTypeNode, macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.ID == "" {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	// Ensure user is authorized ...
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.TagUIDs)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	// Fetch snapshots
	snapshots, err := db.GetSnapshotsForRange(interval, macAddr, periodStartTimestamp, periodEndTimestamp)
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(snapshots)
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

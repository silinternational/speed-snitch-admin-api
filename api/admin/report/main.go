package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/reporting"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const PeriodTimeFormat = "2006-01-02"

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := req.PathParameters["id"]
	if id != "" {
		if strings.HasSuffix(req.Path, "/detail") {
			return getNodeDetailData(req)
		}
		if strings.HasSuffix(req.Path, "/raw") {
			return getNodeRawData(req)
		}
		if strings.HasSuffix(req.Path, "/event") {
			return getNodeReportingEvents(req)
		}
		return viewNodeReport(req)
	}

	return domain.ClientError(http.StatusBadRequest, "id is required in url")
}

func viewNodeReport(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid Node ID")
	}

	// Validate Inputs
	interval := req.QueryStringParameters["interval"]
	if !reporting.IsValidReportingInterval(interval) {
		return domain.ClientError(http.StatusBadRequest, "Invalid interval specified")
	}

	periodStartTimestamp, err := getTimestampFromString(req.QueryStringParameters["start"], "start")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp, err := getTimestampFromString(req.QueryStringParameters["end"], "end")
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

func getNodeDetailData(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid Node ID")
	}

	// Validate Inputs
	periodStartTimestamp, err := getTimestampFromString(req.QueryStringParameters["start"], "start")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp, err := getTimestampFromString(req.QueryStringParameters["end"], "end")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp = periodEndTimestamp + domain.SecondsPerDay - 1

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

	data := []domain.ReportingSnapshot{}

	taskType, errString := getTaskTypeIfValid(req)
	if errString != "" {
		return domain.ClientError(http.StatusBadRequest, errString)
	}

	switch taskType {
	case domain.TaskTypePing:
		data, err = reporting.GetPingLogsAsSnapshots(node, periodStartTimestamp, periodEndTimestamp)
	case domain.TaskTypeSpeedTest:
		data, err = reporting.GetSpeedTestLogsAsSnapshots(node, periodStartTimestamp, periodEndTimestamp)
	case domain.LogTypeRestart:
		data, err = reporting.GetRestartsAsSnapshots(node, periodStartTimestamp, periodEndTimestamp)
	case domain.LogTypeDowntime:
		data, err = reporting.GetNetworkDowntimeAsSnapshots(node, periodStartTimestamp, periodEndTimestamp)
	}

	return domain.ReturnJsonOrError(data, err)
}

func getNodeReportingEvents(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	nodeID := domain.GetResourceIDFromRequest(req)
	if nodeID == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid Node ID")
	}

	// Validate Inputs
	periodStartTimestamp, err := getTimestampFromString(req.QueryStringParameters["start"], "start")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp, err := getTimestampFromString(req.QueryStringParameters["end"], "end")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	// Fetch node to ensure exists and get tags for authorization
	var node domain.Node
	err = db.GetItem(&node, nodeID)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Ensure user is authorized ...
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	// Fetch events
	events, err := db.GetReportingEventsForRange(nodeID, periodStartTimestamp, periodEndTimestamp)
	return domain.ReturnJsonOrError(events, err)
}

func getTaskLogPingTestCSV(node domain.Node, startTimestamp, endTimestamp int64) (events.APIGatewayProxyResponse, error) {
	logItems := []domain.TaskLogPingTest{}
	err := db.GetTaskLogForRange(&logItems, node.ID, startTimestamp, endTimestamp)
	if err != nil {
		err = fmt.Errorf(
			"Error getting ping data for node ID: %v between %v and %v.\n%s",
			node.ID,
			startTimestamp,
			endTimestamp,
			err.Error(),
		)
		return domain.ReturnCSVOrError([]domain.TaskLogMapper{}, "", err)
	}

	// You can't use a slice of structs as a slice of interfaces
	logMappers := make([]domain.TaskLogMapper, len(logItems))
	for i := range logItems {
		logMappers[i] = logItems[i]
	}

	filename := getCSVFilename(node, "ping", startTimestamp, endTimestamp)
	return domain.ReturnCSVOrError(logMappers, filename, nil)

}

func getTaskLogSpeedTestCSV(node domain.Node, startTimestamp, endTimestamp int64) (events.APIGatewayProxyResponse, error) {
	var logItems []domain.TaskLogSpeedTest
	err := db.GetTaskLogForRange(&logItems, node.ID, startTimestamp, endTimestamp)

	if err != nil {
		err = fmt.Errorf(
			"Error getting speed test data for node ID: %v between %v and %v.\n%s",
			node.ID,
			startTimestamp,
			endTimestamp,
			err.Error(),
		)
		return domain.ReturnCSVOrError([]domain.TaskLogMapper{}, "", err)
	}

	// You can't use a slice of structs as a slice of interfaces
	logMappers := make([]domain.TaskLogMapper, len(logItems))
	for i := range logItems {
		logMappers[i] = logItems[i]
	}

	filename := getCSVFilename(node, "speed", startTimestamp, endTimestamp)
	return domain.ReturnCSVOrError(logMappers, filename, nil)
}

func getTaskLogDowntimeCSV(node domain.Node, startTimestamp, endTimestamp int64) (events.APIGatewayProxyResponse, error) {
	var logItems []domain.TaskLogNetworkDowntime
	err := db.GetTaskLogForRange(&logItems, node.ID, startTimestamp, endTimestamp)
	if err != nil {
		err = fmt.Errorf(
			"Error getting network downtime data for node ID: %v between %v and %v.\n%s",
			node.ID,
			startTimestamp,
			endTimestamp,
			err.Error(),
		)
		return domain.ReturnCSVOrError([]domain.TaskLogMapper{}, "", err)
	}
	logMappers := make([]domain.TaskLogMapper, len(logItems))
	for i := range logItems {
		logMappers[i] = logItems[i]
	}

	filename := getCSVFilename(node, "downtime", startTimestamp, endTimestamp)
	return domain.ReturnCSVOrError(logMappers, filename, nil)
}

func getTaskLogRestartCSV(node domain.Node, startTimestamp, endTimestamp int64) (events.APIGatewayProxyResponse, error) {
	var logItems []domain.TaskLogRestart
	err := db.GetTaskLogForRange(&logItems, node.ID, startTimestamp, endTimestamp)
	if err != nil {
		err = fmt.Errorf(
			"Error getting restart data for node ID: %v between %v and %v.\n%s",
			node.ID,
			startTimestamp,
			endTimestamp,
			err.Error(),
		)
		return domain.ReturnCSVOrError([]domain.TaskLogMapper{}, "", err)
	}
	logMappers := make([]domain.TaskLogMapper, len(logItems))
	for i := range logItems {
		logMappers[i] = logItems[i]
	}

	filename := getCSVFilename(node, "restart", startTimestamp, endTimestamp)
	return domain.ReturnCSVOrError(logMappers, filename, nil)
}

func getNodeRawData(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid Node ID")
	}

	taskType, errString := getTaskTypeIfValid(req)
	if errString != "" {
		return domain.ClientError(http.StatusBadRequest, errString)
	}

	// Validate Inputs
	periodStartTimestamp, err := getTimestampFromString(req.QueryStringParameters["start"], "start")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp, err := getTimestampFromString(req.QueryStringParameters["end"], "end")
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	periodEndTimestamp = periodEndTimestamp + domain.SecondsPerDay - 1

	// Fetch node to ensure exists and get tags for authorization
	var node domain.Node
	err = db.GetItem(&node, id)
	if err != nil {
		return domain.ServerError(err)
	}

	// Ensure user is authorized ...
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	switch taskType {
	case domain.TaskTypePing:
		return getTaskLogPingTestCSV(node, periodStartTimestamp, periodEndTimestamp)

	case domain.TaskTypeSpeedTest:
		return getTaskLogSpeedTestCSV(node, periodStartTimestamp, periodEndTimestamp)

	case domain.LogTypeDowntime:
		return getTaskLogDowntimeCSV(node, periodStartTimestamp, periodEndTimestamp)

	case domain.LogTypeRestart:
		return getTaskLogRestartCSV(node, periodStartTimestamp, periodEndTimestamp)

	}

	return domain.ClientError(
		http.StatusBadRequest,
		fmt.Sprintf(`Invalid "type"" query parameter. Must be "%s", "%s", "%s" or "%s". Got: %s.`,
			domain.TaskTypePing, domain.TaskTypeSpeedTest, domain.LogTypeDowntime, domain.LogTypeRestart, taskType),
	)
}

func getTimestampFromString(date, paramName string) (int64, error) {
	errMsg := fmt.Sprintf("%s parameter is required and should be format YYYY-MM-DD", paramName)
	if date == "" {
		return 0, fmt.Errorf(errMsg)
	}
	dateTime, err := time.Parse(PeriodTimeFormat, date)
	if err != nil {
		return 0, fmt.Errorf("%s.\nGot error: %s", errMsg, err.Error())
	}
	timestamp, _, err := reporting.GetStartEndTimestampsForDate(dateTime, "", "")
	if err != nil {
		return 0, fmt.Errorf("%s.\nGot error: %s", errMsg, err.Error())
	}

	return timestamp, nil
}

func getTaskTypeIfValid(req events.APIGatewayProxyRequest) (string, string) {
	taskType := req.QueryStringParameters["type"]
	if taskType != domain.TaskTypePing &&
		taskType != domain.TaskTypeSpeedTest &&
		taskType != domain.LogTypeDowntime &&
		taskType != domain.LogTypeRestart {
		return "", fmt.Sprintf(
			`Invalid "type"" query parameter. Must be "%s", "%s", "%s", or "%s". Got %s.`,
			domain.TaskTypePing, domain.TaskTypeSpeedTest, domain.LogTypeDowntime, domain.LogTypeRestart, taskType,
		)
	}

	return taskType, ""
}

func getCSVFilename(node domain.Node, dataType string, startTimestamp, endTimestamp int64) string {

	// Borrowed this regex stuff from https://github.com/kennygrant/sanitize/blob/master/sanitize.go
	var (
		illegalPath = regexp.MustCompile(`[^[:alnum:]\~\-\./]`)
		separators  = regexp.MustCompile(`[ &_=+:]`)
		dashes      = regexp.MustCompile(`[\-]+`)
	)

	nodeName := node.Nickname
	nodeName = separators.ReplaceAllString(nodeName, "-")
	nodeName = illegalPath.ReplaceAllString(nodeName, " ")
	nodeName = dashes.ReplaceAllString(nodeName, "-")
	nodeName = strings.Trim(nodeName, " ")

	startDate := time.Unix(startTimestamp, 0).UTC().Format(domain.DateLayout)
	endDate := time.Unix(endTimestamp, 0).UTC().Format(domain.DateLayout)

	filename := fmt.Sprintf(`"%s %s from %s to %s.csv"`, dataType, nodeName, startDate, endDate)
	return filename
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

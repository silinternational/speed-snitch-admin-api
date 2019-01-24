package admin

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"strings"
	"testing"
)

func TestViewNodeReport(t *testing.T) {
	testutils.ResetDb(t)

	tagPass := domain.Tag{
		Name: "tag-pass",
	}
	tagFail := domain.Tag{
		Name: "tag-fail",
	}

	tagFixtures := []*domain.Tag{&tagPass, &tagFail}

	for _, i := range tagFixtures {
		db.PutItem(i)
	}

	passNode := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	// Create the node in the database
	err := db.PutItemWithAssociations(
		&passNode,
		[]domain.AssociationReplacements{{Replacements: []domain.Tag{tagPass}, AssociationName: "tags"}},
	)
	if err != nil {
		t.Errorf("Error creating Node fixture.\n%s", err.Error())
	}

	passUser := domain.User{
		UUID:  "pass_test",
		Email: "userPass@test.com",
		Role:  domain.UserRoleAdmin,
	}

	failUser := domain.User{
		UUID:  "fail_test",
		Email: "userFail@test.com",
		Role:  domain.UserRoleAdmin,
	}

	// Create the user in the database
	err = db.PutItemWithAssociations(
		&passUser,
		[]domain.AssociationReplacements{{Replacements: []domain.Tag{tagPass}, AssociationName: "tags"}},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	// Create the user in the database
	err = db.PutItemWithAssociations(
		&failUser,
		[]domain.AssociationReplacements{{Replacements: []domain.Tag{tagFail}, AssociationName: "tags"}},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	taskLogFixtures := []*domain.ReportingSnapshot{
		{
			Interval:    domain.ReportingIntervalDaily,
			Timestamp:   1527811200, // 2018-06-01 00:00:00
			NodeID:      passNode.ID,
			UploadAvg:   20,
			UploadMax:   30,
			UploadMin:   10,
			DownloadAvg: 40,
			DownloadMax: 70,
			DownloadMin: 10,
			LatencyAvg:  4,
			LatencyMax:  6,
			LatencyMin:  2,
		},
		{
			Interval:    domain.ReportingIntervalDaily,
			Timestamp:   1527897600, // 2018-06-02 00:00:00
			NodeID:      passNode.ID,
			UploadAvg:   20,
			UploadMax:   30,
			UploadMin:   10,
			DownloadAvg: 40,
			DownloadMax: 70,
			DownloadMin: 10,
			LatencyAvg:  4,
			LatencyMax:  6,
			LatencyMin:  2,
		},
		{
			Interval:    domain.ReportingIntervalDaily,
			Timestamp:   1527984000, // 2018-06-03 00:00:00
			NodeID:      passNode.ID,
			UploadAvg:   20,
			UploadMax:   30,
			UploadMin:   10,
			DownloadAvg: 40,
			DownloadMax: 70,
			DownloadMin: 10,
			LatencyAvg:  4,
			LatencyMax:  6,
			LatencyMin:  2,
		},
	}
	for _, i := range taskLogFixtures {
		db.PutItem(i)
	}

	strNodeID := fmt.Sprintf("%d", passNode.ID)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/" + strNodeID,
		PathParameters: map[string]string{
			"id": strNodeID,
		},
		Headers: map[string]string{
			"x-user-uuid": passUser.UUID,
			"x-user-mail": passUser.Email,
		},
		QueryStringParameters: map[string]string{
			"interval": "daily",
			"start":    "2018-06-01",
			"end":      "2018-06-03",
		},
	}

	response, err := viewNodeReport(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	var snapResults []domain.ReportingSnapshot
	err = json.Unmarshal([]byte(response.Body), &snapResults)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if len(snapResults) != len(taskLogFixtures) {
		t.Error("Not correct number of snapshots returned. Expected", len(taskLogFixtures), "got", len(snapResults))
		t.Fail()
	}

	// try again with user who is not allowed to view this node to ensure error message
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/" + strNodeID,
		PathParameters: map[string]string{
			"id": strNodeID,
		},
		Headers: map[string]string{
			"x-user-uuid": failUser.UUID,
			"x-user-mail": failUser.Email,
		},
		QueryStringParameters: map[string]string{
			"interval": "daily",
			"start":    "2018-06-01",
			"end":      "2018-06-03",
		},
	}

	response, err = viewNodeReport(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 403 {
		t.Error("Wrong status code returned, expected 403, got", response.StatusCode, response.Body)
	}
}

func getRawDataRequest(nodeID, logType, startDate, endDate string) events.APIGatewayProxyRequest {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/nodeID/raw",
		PathParameters: map[string]string{
			"id": nodeID,
		},
		Headers: map[string]string{
			"x-user-uuid": testutils.SuperAdmin.UUID,
			"x-user-mail": testutils.SuperAdmin.Email,
		},
		QueryStringParameters: map[string]string{
			"type":  logType,
			"start": startDate,
			"end":   endDate,
		},
	}

	return req
}

func isCSVValid(csvString string) error {
	// Check the csv is OK
	ioResults := strings.NewReader(csvString)
	reader := csv.NewReader(ioResults)

	_, err := reader.ReadAll()
	return err
}

func TestGetNodeRawData(t *testing.T) {
	testutils.ResetDb(t)

	tagPass := domain.Tag{
		Name: "tag-pass",
	}
	tagFail := domain.Tag{
		Name: "tag-fail",
	}

	tagFixtures := []*domain.Tag{&tagPass, &tagFail}

	for _, i := range tagFixtures {
		db.PutItem(i)
	}

	passNode := domain.Node{
		MacAddr:  "aa:aa:aa:aa:aa:aa",
		Nickname: "Africa(test)",
	}

	// Create the node in the database
	err := db.PutItemWithAssociations(
		&passNode,
		[]domain.AssociationReplacements{{Replacements: []domain.Tag{tagPass}, AssociationName: "tags"}},
	)
	if err != nil {
		t.Errorf("Error creating Node fixture.\n%s", err.Error())
	}

	strPassNodeID := fmt.Sprintf("%d", passNode.ID)

	failNode := domain.Node{
		MacAddr: "21:22:23:24:25:26",
	}

	db.PutItem(&failNode)

	speedInRange := []*domain.TaskLogSpeedTest{
		{
			NodeID:    passNode.ID,
			Timestamp: 1528145185,
			Upload:    10.0,
			Download:  10.0,
		},
		{
			NodeID:    passNode.ID,
			Timestamp: 1528145285,
			Upload:    20.0,
			Download:  20.0,
		},
		{
			NodeID:    failNode.ID,
			Timestamp: 1528145385,
			Upload:    30.0,
			Download:  30.0,
		},
		{
			NodeID:    passNode.ID,
			Timestamp: 1528160000,
			Upload:    40.0,
			Download:  40.0,
		},
	}

	for _, i := range speedInRange {
		db.PutItem(i)
	}

	pingInRange := []*domain.TaskLogPingTest{
		{
			NodeID:            passNode.ID,
			Timestamp:         1528145485,
			Latency:           5,
			PacketLossPercent: 5,
		},
		{
			NodeID:            passNode.ID,
			Timestamp:         1528145486,
			Latency:           10,
			PacketLossPercent: 10,
		},
		{
			NodeID:            failNode.ID,
			Timestamp:         1528145489,
			Latency:           15,
			PacketLossPercent: 15,
		},
	}

	for _, i := range pingInRange {
		db.PutItem(i)
	}

	downtimeInRange := []*domain.TaskLogNetworkDowntime{
		{
			NodeID:          passNode.ID,
			Timestamp:       1528145000,
			DowntimeSeconds: 111,
		},
		{
			NodeID:          passNode.ID,
			Timestamp:       1528146000,
			DowntimeSeconds: 222,
		},
		{
			NodeID:          failNode.ID,
			Timestamp:       1528145489,
			DowntimeSeconds: 333,
		},
	}

	for _, i := range downtimeInRange {
		db.PutItem(i)
	}

	restartsInRange := []*domain.TaskLogRestart{
		{
			NodeID:    passNode.ID,
			Timestamp: 1528145490,
		},
		{
			NodeID:    passNode.ID,
			Timestamp: 1528145491,
		},
		{
			NodeID:    failNode.ID,
			Timestamp: 1528145491,
		},
	}
	for _, i := range restartsInRange {
		db.PutItem(i)
	}

	// Test for passNode's speedTest logs
	response, err := getNodeRawData(getRawDataRequest(strPassNodeID, domain.TaskTypeSpeedTest, "2018-06-04", "2018-06-05"))
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results := response.Body
	if !strings.Contains(results, `,10.000,`) ||
		!strings.Contains(results, `,20.000,`) ||
		!strings.Contains(results, `,40.000,`) ||
		// should not have this one
		strings.Contains(results, `30.000`) {
		t.Errorf("Expected three logs with values of 10.000, 20.000 and 40.000, but got\n%s", results)
	}

	// Test for passNode's ping logs
	response, err = getNodeRawData(getRawDataRequest(strPassNodeID, domain.TaskTypePing, "2018-06-04", "2018-06-04"))
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results = response.Body
	if !strings.Contains(results, `,5.000,`) ||
		!strings.Contains(results, `,10.000,`) ||
		strings.Contains(results, `15`) {
		t.Errorf("Expected two logs with values of 5 and 10, but got\n%s", results)
	}

	// Check the csv is OK
	err = isCSVValid(results)
	if err != nil {
		t.Errorf("Error reading results as csv.\n %s", err.Error())
		return
	}

	contType, ok := response.Headers["Content-Type"]
	expectedCT := "text/csv"

	if !ok {
		t.Errorf("Missing Content Type. \nExpected: %s ", expectedCT)
		return
	}
	if expectedCT != contType {
		t.Errorf("Wrong Content Type. \nExpected: %s \n But Got: %s", expectedCT, contType)
		return
	}

	contDisposition, ok := response.Headers["Content-Disposition"]
	expectedCD := `attachment;filename="ping Africa test from 2018-06-04 to 2018-06-04.csv"`

	if !ok {
		t.Errorf("Missing Content Disposition. \nExpected: %s ", expectedCD)
		return
	}
	if expectedCD != contDisposition {
		t.Errorf("Wrong Content Disposition. \nExpected: %s \n But Got: %s", expectedCD, contDisposition)
		return
	}

	// Test for passNode's downtime logs
	response, err = getNodeRawData(getRawDataRequest(strPassNodeID, domain.LogTypeDowntime, "2018-06-04", "2018-06-04"))
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results = response.Body
	if !strings.Contains(results, `,111,`) ||
		!strings.Contains(results, `,222,`) ||
		strings.Contains(results, `333`) {
		t.Errorf("Expected two logs with values of 111 and 222, but got\n%s", results)
	}

	// Check the csv is OK
	err = isCSVValid(results)
	if err != nil {
		t.Errorf("Error reading results as csv.\n %s", err.Error())
		return
	}

	// Test for passNode's Restart logs
	response, err = getNodeRawData(getRawDataRequest(strPassNodeID, domain.LogTypeRestart, "2018-06-04", "2018-06-04"))
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results = response.Body
	if !strings.Contains(results, `:30`) ||
		!strings.Contains(results, `:31`) ||
		strings.Contains(results, `2,2018`) {
		t.Errorf("Expected two logs with values of 30 and 31, but got\n%s", results)
		return
	}

	err = isCSVValid(results)
	if err != nil {
		t.Errorf("Error reading results as csv.\n %s", err.Error())
		return
	}

}

func TestGetReportingEventsForRange(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}
	db.PutItem(&node1)

	node2 := domain.Node{
		MacAddr: "bb:bb:bb:bb:bb:bb",
	}
	db.PutItem(&node2)

	june4 := int64(1528070400)
	june5 := int64(1528156800)

	eventsInRange := []domain.ReportingEvent{
		{
			Timestamp: june4,
			Name:      "No Node June 4th",
		},
		{
			NodeID:    node1.ID,
			Timestamp: june4,
			Name:      "Node1 June 4th",
		},
		{
			NodeID:    node1.ID,
			Timestamp: june5,
			Name:      "Node1 June 5th",
		},
		{
			NodeID:    node2.ID,
			Timestamp: june5,
			Name:      "Node2 June 5th",
		},
	}

	for _, i := range eventsInRange {
		db.PutItem(&i)
	}

	eventsOutOfRange := []domain.ReportingEvent{
		{
			Timestamp: 1527984000,
			Name:      "No Node June 3rd",
		},
		{
			NodeID:    node1.ID,
			Timestamp: 1528243200,
			Name:      "Node1 June 6th",
		},
	}

	for _, i := range eventsOutOfRange {
		db.PutItem(&i)
	}

	strNodeID := fmt.Sprintf("%d", node1.ID)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/" + strNodeID + "/events",
		PathParameters: map[string]string{
			"id": strNodeID,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
		QueryStringParameters: map[string]string{
			"start": "2018-06-04",
			"end":   "2018-06-05",
		},
	}

	response, err := getNodeReportingEvents(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	var eventResults []domain.ReportingEvent
	err = json.Unmarshal([]byte(response.Body), &eventResults)
	if err != nil {
		t.Error(err)
		return
	}

	expectedLen := len(eventsInRange) - 1
	if len(eventResults) != expectedLen {
		t.Errorf("Wrong number of Reporting Events returned. Expected: %d. Got: %d", expectedLen, len(eventResults))
		return
	}

	for _, event := range eventResults {
		if event.NodeID != 0 && event.NodeID != node1.ID {
			t.Errorf("Got an unexpected event in the results. \n%+v", event)
		}
		if event.Timestamp != june4 && event.Timestamp != june5 {
			t.Errorf("Got an unexpected event in the results. \n%+v", event)
		}
	}
}

package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"testing"
)

func TestViewNodeReport(t *testing.T) {
	db.FlushTables(t)

	nodeFixtures := []domain.Node{
		{
			ID:      "node-aa:aa:aa:aa:aa:aa",
			MacAddr: "aa:aa:aa:aa:aa:aa",
			TagUIDs: []string{"pass"},
		},
	}

	userFixtures := []domain.User{
		{
			ID:      "user-pass",
			UID:     "pass",
			UserID:  "pass_test",
			Role:    domain.UserRoleAdmin,
			TagUIDs: []string{"pass"},
		},
		{
			ID:      "user-fail",
			UID:     "fail",
			UserID:  "fail_test",
			Role:    domain.UserRoleAdmin,
			TagUIDs: []string{"fail"},
		},
	}

	taskLogFixtures := []domain.ReportingSnapshot{
		{
			ID:          "daily-aa:aa:aa:aa:aa:aa",
			Timestamp:   1527811200, // 2018-06-01 00:00:00
			MacAddr:     "aa:aa:aa:aa:aa:aa",
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
			ID:          "daily-aa:aa:aa:aa:aa:aa",
			Timestamp:   1527897600, // 2018-06-02 00:00:00
			MacAddr:     "aa:aa:aa:aa:aa:aa",
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
			ID:          "daily-aa:aa:aa:aa:aa:aa",
			Timestamp:   1527984000, // 2018-06-03 00:00:00
			MacAddr:     "aa:aa:aa:aa:aa:aa",
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

	for _, fix := range nodeFixtures {
		err := db.PutItem(domain.DataTable, fix)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	for _, fix := range userFixtures {
		err := db.PutItem(domain.DataTable, fix)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	for _, fix := range taskLogFixtures {
		err := db.PutItem(domain.TaskLogTable, fix)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/aa:aa:aa:aa:aa:aa",
		PathParameters: map[string]string{
			"macAddr": "aa:aa:aa:aa:aa:aa",
		},
		Headers: map[string]string{
			"x-user-id": "pass_test",
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
		Path:       "/report/node/aa:aa:aa:aa:aa:aa",
		PathParameters: map[string]string{
			"macAddr": "aa:aa:aa:aa:aa:aa",
		},
		Headers: map[string]string{
			"x-user-id": "fail_test",
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

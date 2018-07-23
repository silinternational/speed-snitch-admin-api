package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"testing"
)

func TestViewNodeReport(t *testing.T) {
	testutils.ResetDb(t)

	tagPass := domain.Tag{
		Model: gorm.Model{
			ID: 1,
		},
		Name: "tag-pass",
	}
	tagFail := domain.Tag{
		Model: gorm.Model{
			ID: 2,
		},
		Name: "tag-fail",
	}

	tagFixtures := []domain.Tag{tagPass, tagFail}

	for _, i := range tagFixtures {
		db.PutItem(&i)
	}

	passNode := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	// Create the node in the database
	err := db.PutItemWithAssociations(
		&passNode,
		[]domain.AssociationReplacement{{Replacement: tagPass, AssociationName: "tags"}},
	)

	passUser := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "pass_test",
		Email: "userPass@test.com",
		Role:  domain.UserRoleAdmin,
	}

	failUser := domain.User{
		Model: gorm.Model{
			ID: 3,
		},
		UUID:  "fail_test",
		Email: "userFail@test.com",
		Role:  domain.UserRoleAdmin,
	}

	// Create the user in the database
	err = db.PutItemWithAssociations(
		&passUser,
		[]domain.AssociationReplacement{{Replacement: tagPass, AssociationName: "tags"}},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	// Create the user in the database
	err = db.PutItemWithAssociations(
		&failUser,
		[]domain.AssociationReplacement{{Replacement: tagFail, AssociationName: "tags"}},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	taskLogFixtures := []domain.ReportingSnapshot{
		{
			Model: gorm.Model{
				ID: 1,
			},
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
			Model: gorm.Model{
				ID: 2,
			},
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
			Model: gorm.Model{
				ID: 3,
			},
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
		db.PutItem(&i)
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/1",
		PathParameters: map[string]string{
			"id": "1",
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
		Path:       "/report/node/1",
		PathParameters: map[string]string{
			"id": "1",
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

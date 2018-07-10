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

func TestHandler(t *testing.T) {
	testutils.ResetDb(t)

	version1 := domain.Version{
		Model: gorm.Model{
			ID: 1,
		},
		Number:      "1.1.1",
		Description: "Version 1",
	}

	version2 := domain.Version{
		Model: gorm.Model{
			ID: 2,
		},
		Number:      "2.2.2",
		Description: "Version 2",
	}

	versionFixtures := []domain.Version{version1, version2}

	for _, fix := range versionFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	node1 := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr:           "11.22.33.44.55.66",
		OS:                "linux",
		Arch:              "arn",
		RunningVersion:    version1,
		ConfiguredVersion: version1,
		Nickname:          "Node Fixture",
	}

	err := db.PutItem(&node1)
	if err != nil {
		t.Error(err)
		return
	}

	method := "POST"

	// Test using a new version for an existing node

	helloReq := domain.HelloRequest{
		ID:      node1.MacAddr,
		Version: version2.Number, // new version (not version1)
		Uptime:  111,
		OS:      node1.OS,
		Arch:    node1.Arch,
	}
	js, err := json.Marshal(helloReq)
	if err != nil {
		t.Error("Unable to marshal hello request to JSON, err: ", err.Error())
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod:     method,
		Path:           "/hello",
		PathParameters: map[string]string{},
		Body:           string(js),
	}

	response, err := Handler(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 204 {
		t.Error("Wrong status code returned, expected 204, got", response.StatusCode, response.Body)
		return
	}

	// fetch node from db to check for update
	var node domain.Node
	err = db.GetItem(&node, node1.Model.ID)
	if err != nil {
		t.Error("Unable to get node, err: ", err.Error())
		return
	}

	if node.RunningVersion.ID != version2.ID {
		t.Errorf("Wrong 'RunningVersion' on node. Expected ID: %d\n But got: %+v.", version2.ID, node.RunningVersion)
		return
	}

	// Test using a missing version for a new node
	// NOTE: This should produce an error log message even though the test doesn't fail.
	helloReq = domain.HelloRequest{
		ID:      "66.55.44.33.22.11", // not in db
		Version: "9.9.9",             // missing version
		Uptime:  999,
		OS:      "darwin",
		Arch:    "amd64",
	}
	js, err = json.Marshal(helloReq)
	if err != nil {
		t.Error("Unable to marshal hello request to JSON, err: ", err.Error())
	}

	req = events.APIGatewayProxyRequest{
		HTTPMethod:     method,
		Path:           "/hello",
		PathParameters: map[string]string{},
		Body:           string(js),
	}

	response, err = Handler(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 204 {
		t.Error("Wrong status code returned, expected 204, got", response.StatusCode, response.Body)
		return
	}

	node = domain.Node{
		MacAddr: helloReq.ID,
	}

	err = db.FindOne(&node)
	if err != nil {
		t.Error("Error retrieving new node: ", err.Error())
		return
	}

	if node.RunningVersionID > 0 {
		t.Errorf("Node created with a 'RunningVersion', but expected it to be empty.\n\tNode: %+v", node)
		return
	}
}

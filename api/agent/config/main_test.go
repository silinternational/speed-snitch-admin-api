package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"strings"
	"testing"
)

func listVersionsCheckLength(expectedLength int) ([]domain.Version, error) {
	versions := []domain.Version{}
	err := db.ListItems(&versions, "id asc")
	if err != nil {
		return versions, fmt.Errorf("Error calling list versions: %s", err.Error())
	}
	if len(versions) != expectedLength {
		return versions, fmt.Errorf("Wrong number of versions. Expected: %d. But got: %d", expectedLength, len(versions))
	}
	return versions, nil
}

func listNodesCheckLength(expectedLength int) ([]domain.Node, error) {
	nodes := []domain.Node{}
	err := db.ListItems(&nodes, "id asc")
	if err != nil {
		return nodes, fmt.Errorf("Error calling list nodes: %s", err.Error())
	}
	if len(nodes) != expectedLength {
		return nodes, fmt.Errorf("Wrong number of nodes. Expected: %d. But got: %d.", expectedLength, len(nodes))
	}
	return nodes, nil
}

func TestGetConfig(t *testing.T) {
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

	task1 := domain.Task{
		Model: gorm.Model{
			ID: 1,
		},
		Type:       domain.TaskTypePing,
		Schedule:   "*/5 * * * *",
		ServerHost: "www.google.com",
	}

	err := db.PutItem(&task1)
	if err != nil {
		t.Error(err)
		return
	}

	node1 := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr:           "11:12:13:14:15:16",
		OS:                "linux",
		Arch:              "arn",
		RunningVersion:    version1,
		ConfiguredVersion: version1,
		Nickname:          "Node Fixture",
		Tasks:             []domain.Task{task1},
	}

	node2 := domain.Node{
		Model: gorm.Model{
			ID: 2,
		},
		MacAddr:           "21:22:23:24:25:26",
		OS:                "linux",
		Arch:              "arn",
		RunningVersion:    version2,
		ConfiguredVersion: version2,
		Nickname:          "Node Fixture2",
	}

	nodeFixtures := []domain.Node{node1, node2}

	for _, fix := range nodeFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	_, err = listVersionsCheckLength(2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	_, err = listNodesCheckLength(2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	method := "GET"

	req := events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/config",
		PathParameters: map[string]string{
			"macAddr": node1.MacAddr,
		},
	}

	response, err := getConfig(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}
	results := response.Body
	if !strings.Contains(results, node1.ConfiguredVersion.Number) || !strings.Contains(results, task1.ServerHost) {
		t.Errorf("getConfig did not include the right data. Got:\n%s\n", results)
	}
}

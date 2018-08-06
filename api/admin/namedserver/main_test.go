package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"net/http"
	"strings"
	"testing"
)

func TestDeleteServer(t *testing.T) {
	testutils.ResetDb(t)

	// Create server record that stays
	toKeepServer := domain.NamedServer{
		Name:       "KeepMe server",
		ServerType: "ping",
		ServerHost: "keep.example.org",
	}
	err := db.PutItem(&toKeepServer)
	if err != nil {
		t.Error("Got error trying to create test record: ", err.Error())
	}

	// Create server record to delete
	toDeleteServer := domain.NamedServer{
		Name:       "DeleteMe server",
		ServerType: "ping",
		ServerHost: "delete.example.org",
	}
	err = db.PutItem(&toDeleteServer)
	if err != nil {
		t.Error("Got error trying to create test record: ", err.Error())
	}

	version1 := domain.Version{
		Number: "1.0.0",
	}
	err = db.PutItem(&version1)
	if err != nil {
		t.Error(err)
		return
	}

	// Create a node with a task that is related to a NamedServer
	task1 := domain.Task{
		Type:          domain.TaskTypePing,
		NamedServerID: toKeepServer.ID,
		Schedule:      "* * * * *",
		ServerHost:    toKeepServer.ServerHost,
	}

	relatedNode := domain.Node{
		MacAddr:             "aa:aa:aa:aa:aa:aa",
		RunningVersionID:    version1.ID,
		ConfiguredVersionID: version1.ID, // This one should be used
		ConfiguredVersion:   version1,    // This one should be ignored
		Tasks:               []domain.Task{task1},
	}
	err = db.PutItem(&relatedNode)
	if err != nil {
		t.Error(err)
		return
	}

	// Call deleteServer
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       fmt.Sprintf("/namedserver/%v", toDeleteServer.ID),
		PathParameters: map[string]string{
			"id": fmt.Sprintf("%v", toDeleteServer.ID),
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := deleteServer(req)
	if err != nil {
		t.Error("Got error trying to delete newly created server: ", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response for delete call, got ", resp.StatusCode)
	}

	// Make sure record was actually deleted
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       fmt.Sprintf("/namedserver/%v", toDeleteServer.ID),
		PathParameters: map[string]string{
			"id": fmt.Sprintf("%v", toDeleteServer.ID),
		},
		Headers: map[string]string{
			"x-user-id": "super_admin",
		},
	}

	resp, err = viewServer(req)
	if err != nil {
		t.Error("Got error trying to check on the deleted server: ", err.Error())
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Should have gotten 404 after deletion but got status: ", resp.StatusCode)
	}

	// Make sure only one NamedServer record remains
	var allNamedServers []domain.NamedServer
	err = db.ListItems(&allNamedServers, "")
	if err != nil {
		t.Errorf("Error trying to check fixture. %s", err.Error())
		return
	}

	if len(allNamedServers) != 1 {
		t.Errorf("Wrong number of namedservers remaining.  Expected 1, but got %d.", len(allNamedServers))
		return
	}

	if allNamedServers[0].Name != toKeepServer.Name {
		t.Errorf(
			"Wrong namedserver was kept. Expected Name: %s. But got Name:%s",
			toKeepServer.Name,
			allNamedServers[0].Name,
		)
	}

	// Try to delete the NamedServer that has a related task. Make sure we get a 409.
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       fmt.Sprintf("/namedserver/%v", toKeepServer.ID),
		PathParameters: map[string]string{
			"id": fmt.Sprintf("%v", toKeepServer.ID),
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err = deleteServer(req)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Did not get expected http status (%v). Got: %v", http.StatusConflict, resp.StatusCode)
	}
	if !strings.Contains(resp.Body, RelatedTaskErrorMessage) {
		t.Errorf("Did not get the expected response. \nExpected it to include: %s\n But got: %s", RelatedTaskErrorMessage, resp.Body)
	}

}

func TestListServers(t *testing.T) {
	testutils.ResetDb(t)

	// First, test that with no servers the response json is an empty array
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/namedserver",
		Headers:    testutils.GetSuperAdminReqHeader(),
	}

	resp, err := listServers(req)
	if err != nil {
		t.Error("Got error trying to list servers: ", err.Error())
	}
	if resp.Body != "[]" {
		t.Error("Response body was not empty array, got: ", resp.Body)
	}

	serverPing := domain.NamedServer{
		Name:       "server1",
		ServerType: domain.ServerTypePing,
	}

	serverSTN := domain.NamedServer{
		Name:       "server2",
		ServerType: domain.ServerTypeSpeedTestNet,
	}

	// Create a couple servers
	servers := []domain.NamedServer{serverPing, serverSTN}

	for _, srv := range servers {
		err = db.PutItem(&srv)
		if err != nil {
			t.Error("Unable to create server for test, got error: ", err.Error())
			return
		}
	}

	// Call API to get list of servers
	resp, err = listServers(req)
	if err != nil {
		t.Error("Got error trying to list servers: ", err.Error())
		return
	}

	var returnedServers []domain.NamedServer
	err = json.Unmarshal([]byte(resp.Body), &returnedServers)

	if len(returnedServers) != len(servers) {
		t.Error("Did not return same number of servers as expected. Got: ", len(returnedServers), " expected: ", len(servers))
		return
	}

	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/namedserver",
		QueryStringParameters: map[string]string{"type": domain.ServerTypeSpeedTestNet},
		Headers:               testutils.GetSuperAdminReqHeader(),
	}

	// Call API to get filtered list of servers
	resp, err = listServers(req)
	if err != nil {
		t.Error("Got error trying to get filtered list of servers: ", err.Error())
		return
	}

	err = json.Unmarshal([]byte(resp.Body), &returnedServers)
	if len(returnedServers) != 1 || returnedServers[0].Name != serverSTN.Name {
		t.Errorf("Wrong named servers returned. \nExpected:\n[%+v]\nBut got:\n%+v", serverSTN, returnedServers)
		return
	}

	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/namedserver",
		QueryStringParameters: map[string]string{"type": "not_a_type"},
		Headers:               testutils.GetSuperAdminReqHeader(),
	}

	// Call API to get filtered list of servers
	resp, err = listServers(req)
	if err != nil {
		t.Error("Got error trying to get filtered list of servers: ", err.Error())
		return
	}

	if resp.Body != "[]" {
		t.Errorf("Wrong response. Expected an empty list of named servers, but got %s", resp.Body)
		return
	}
}

func TestUpdateServer(t *testing.T) {
	testutils.ResetDb(t)

	// Create server record to update
	createServer := domain.NamedServer{
		Name:       "test server",
		ServerType: domain.ServerTypePing,
	}

	err := db.PutItem(&createServer)
	if err != nil {
		t.Error("Got error trying to create test record: ", err.Error())
	}

	createServer.Name = "updated test server"
	js, err := json.Marshal(createServer)
	if err != nil {
		t.Error("Got error trying to marshal json for update: ", err.Error())
	}

	// Call API to update record
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       fmt.Sprintf("/namedserver/%v", createServer.ID),
		PathParameters: map[string]string{
			"id": fmt.Sprintf("%v", createServer.ID),
		},
		Headers: testutils.GetSuperAdminReqHeader(),
		Body:    string(js),
	}

	resp, err := updateServer(req)
	if err != nil {
		t.Error("Got error trying to update test record: ", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected status code 200 for update server, got: ", resp.StatusCode)
	}

	// Find server in db and check value was changed
	var updatedServer domain.NamedServer
	err = db.GetItem(&updatedServer, createServer.ID)
	if err != nil {
		t.Error("Got error trying to get updated test record: ", err.Error())
	}
	if updatedServer.Name != createServer.Name {
		t.Errorf("Doesn't look like server was updated, after update: %+v", updatedServer)
	}
}

func TestUpdateServerFailUniqueName(t *testing.T) {
	testutils.ResetDb(t)

	// Create server fixtures
	server1 := domain.NamedServer{
		Name:       "test server1",
		ServerType: domain.ServerTypePing,
	}

	err := db.PutItem(&server1)
	if err != nil {
		t.Error("Got error trying to load fixture: ", err.Error())
	}

	server2 := domain.NamedServer{
		Name:       "test server2",
		ServerType: domain.ServerTypePing,
	}

	err = db.PutItem(&server2)
	if err != nil {
		t.Error("Got error trying to load fixture: ", err.Error())
	}

	// New Test Server should fail using same name as a fixture
	newServer := domain.NamedServer{
		Name:       server1.Name,
		ServerType: domain.ServerTypePing,
	}

	js, err := json.Marshal(&newServer)
	if err != nil {
		t.Error("Got error trying to marshal json for new server: ", err.Error())
	}

	// Call API to update record
	req := events.APIGatewayProxyRequest{
		HTTPMethod:     "PUT",
		Path:           fmt.Sprintf("/namedserver"),
		PathParameters: map[string]string{},
		Headers:        testutils.GetSuperAdminReqHeader(),
		Body:           string(js),
	}

	resp, err := updateServer(req)
	if err != nil {
		t.Error("Got error trying to update test record: ", err.Error())
		return
	}
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status code %v for update server, got: %v", http.StatusConflict, resp.StatusCode)
	}

	if !strings.Contains(resp.Body, UniqueNameErrorMessage) {
		t.Errorf("Did not get the expected response. \nExpected it to include: %s\n But got: %s", UniqueNameErrorMessage, resp.Body)
	}

	// Updating an existing Server should fail using same name as the other fixture
	server1.Name = server2.Name

	js, err = json.Marshal(&server1)
	if err != nil {
		t.Error("Got error trying to marshal json for new server: ", err.Error())
	}

	// Call API to update record
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       fmt.Sprintf("/namedserver/%v", server1.ID),
		PathParameters: map[string]string{
			"id": fmt.Sprintf("%v", server1.ID),
		},
		Headers: testutils.GetSuperAdminReqHeader(),
		Body:    string(js),
	}

	resp, err = updateServer(req)
	if err != nil {
		t.Error("Got error trying to update test record: ", err.Error())
		return
	}
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status code %v for update server, got: %v", http.StatusConflict, resp.StatusCode)
	}

	if !strings.Contains(resp.Body, UniqueNameErrorMessage) {
		t.Errorf("Did not get the expected response. \nExpected it to include: %s\n But got: %s", UniqueNameErrorMessage, resp.Body)
	}
}

func TestViewServer(t *testing.T) {
	testutils.ResetDb(t)

	// Test error 400 if id is missing in path parameters
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/namedserver",
		Headers:    testutils.GetSuperAdminReqHeader(),
	}

	resp, err := viewServer(req)
	if err != nil {
		t.Error("Got error trying to view server without id: ", err.Error())
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Did not get back 400 error for view server call without id, got: ", resp.StatusCode)
	}

	// Next test a server that doesn't exist to ensure 404 error
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/namedserver/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err = viewServer(req)
	if err != nil {
		t.Error("Got error trying to view server: ", err.Error())
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Did not get back 404 error for server that doesnt exist, got: ", resp.StatusCode)
	}

	// Create record to search for
	createServer := domain.NamedServer{
		Name: "test server",
	}
	err = db.PutItem(&createServer)
	if err != nil {
		t.Error("Got error trying to create test record: ", err.Error())
	}

	// Now call viewServer again to try to get newly created record
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       fmt.Sprintf("/namedserver/%v", createServer.ID),
		PathParameters: map[string]string{
			"id": fmt.Sprintf("%v", createServer.ID),
		},
		Headers: map[string]string{
			"x-user-id": "super_admin",
		},
	}

	resp, err = viewServer(req)
	if err != nil {
		t.Error("Got error trying to view newly created server: ", err.Error())
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Error("Got back error 404 for record that should exist, id: ", createServer.ID, "response: ", resp.Body)
	}

}

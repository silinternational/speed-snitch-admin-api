package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"testing"
)

const TestHostForSpeedTestNet = "SpeedTestNetFixtureHost"
const TestServerIDForSpeedTestNet = "111"

func getCustomNamedServerFixtures(uid, serverHost string) []domain.NamedServer {
	namedServerFixtures := []domain.NamedServer{
		{
			ID:         domain.DataTypeNamedServer + "-" + uid,
			UID:        uid,
			ServerType: domain.ServerTypeCustom,
			ServerHost: serverHost,
		},
	}
	return namedServerFixtures
}

func getNodeFixtures() []domain.Node {
	nodeFixtures := []domain.Node{
		{
			ID:      "node-aa:aa:aa:aa:aa:aa",
			Arch:    "arm",
			MacAddr: "aa:aa:aa:aa:aa:aa",
			Tags:    []domain.Tag{getTagFixtures()[0]},
		},
	}
	return nodeFixtures
}

func getTagFixtures() []domain.Tag {

	tagFixtures := []domain.Tag{
		{ID: "tag-pass", UID: "pass", Name: "Pass"},
		{ID: "tag-fail", UID: "fail", Name: "Fail"},
	}
	return tagFixtures
}

func getUserFixtures() []domain.User {
	tagFixtures := getTagFixtures()

	userFixtures := []domain.User{
		{
			ID:     "user-superadmin",
			UID:    "superadmin",
			UserID: "super_admin",
			Role:   domain.UserRoleSuperAdmin,
		},
		{
			ID:     "user-pass",
			UID:    "pass",
			UserID: "pass_test",
			Role:   domain.UserRoleAdmin,
			Tags:   []domain.Tag{tagFixtures[0]},
		},
		{
			ID:     "user-fail",
			UID:    "fail",
			UserID: "fail_test",
			Role:   domain.UserRoleAdmin,
			Tags:   []domain.Tag{tagFixtures[1]},
		},
	}

	return userFixtures
}

func areStringMapsEqual(expected, results map[string]string) bool {
	if len(expected) != len(results) {
		return false
	}

	for key, expectedValue := range expected {
		resultValue, ok := results[key]
		if !ok {
			return false
		}
		if resultValue != expectedValue {
			return false
		}
	}

	return true
}

func areIntMapsEqual(expected, results map[string]int) bool {
	if len(expected) != len(results) {
		return false
	}

	for key, expectedValue := range expected {
		resultValue, ok := results[key]
		if !ok {
			return false
		}
		if resultValue != expectedValue {
			return false
		}
	}

	return true
}

func areIntSliceMapsEqual(expected, results map[string][]int) bool {
	if len(expected) != len(results) {
		return false
	}

	for key, expectedValue := range expected {
		resultValue, ok := results[key]
		if !ok {
			return false
		}
		if len(resultValue) != len(expectedValue) {
			return false
		}
		for index, nextExpected := range expectedValue {
			if resultValue[index] != nextExpected {
				return false
			}
		}
	}

	return true
}

func areFloatMapsEqual(expected, results map[string]float64) bool {
	if len(expected) != len(results) {
		return false
	}

	for key, expectedValue := range expected {
		resultValue, ok := results[key]
		if !ok {
			return false
		}
		if resultValue != expectedValue {
			return false
		}
	}

	return true
}

func TestGetPingStringValuesWithoutNamedServer(t *testing.T) {
	task := domain.Task{}
	task.NamedServer = domain.NamedServer{}
	results, err := getPingStringValues(task)
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: domain.DefaultPingServerHost,
		ServerIDKey:   domain.DefaultPingServerID,
	}

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
}

func TestGetPingStringValuesWithNamedServer(t *testing.T) {

	serverHost := "PingTestHost"
	namedServerUID := "ns11"

	namedServerFixtures := getCustomNamedServerFixtures(namedServerUID, serverHost)
	db.LoadNamedServerFixtures(namedServerFixtures, t)

	task := domain.Task{}
	task.NamedServer = namedServerFixtures[0]

	results, err := getPingStringValues(task)
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: serverHost,
		ServerIDKey:   namedServerUID,
	}

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
}

func TestUpdateTaskPingWithoutNamedServer(t *testing.T) {
	task := domain.Task{}
	task.NamedServer = domain.NamedServer{}

	resultsTask, err := updateTaskPing(task)

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: domain.DefaultPingServerHost,
		ServerIDKey:   domain.DefaultPingServerID,
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}

	resultsInts := resultsTask.Data.IntValues
	expectedInts := map[string]int{TimeOutKey: DefaultPingTimeoutInSeconds}

	if !areIntMapsEqual(expectedInts, resultsInts) {
		t.Errorf("Bad IntValues.\nExpected: %v.\n But got: %v", expectedInts, resultsInts)
	}
}

func TestUpdateTaskPingWithNamedServer(t *testing.T) {
	serverHost := "PingTestHost"
	namedServerUID := "nst12"

	namedServerFixtures := getCustomNamedServerFixtures(namedServerUID, serverHost)
	db.LoadNamedServerFixtures(namedServerFixtures, t)

	task := domain.Task{}
	task.NamedServer = namedServerFixtures[0]

	resultsTask, err := updateTaskPing(task)

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: serverHost,
		ServerIDKey:   namedServerUID,
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}

	resultsInts := resultsTask.Data.IntValues
	expectedInts := map[string]int{TimeOutKey: DefaultPingTimeoutInSeconds}

	if !areIntMapsEqual(expectedInts, resultsInts) {
		t.Errorf("Bad IntValues.\nExpected: %v.\n But got: %v", expectedInts, resultsInts)
	}
}

func TestGetSpeedTestStringValuesWithoutNamedServer(t *testing.T) {
	task := domain.Task{}
	task.NamedServer = domain.NamedServer{}

	results, err := getSpeedTestStringValues(task)
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: domain.DefaultSpeedTestNetServerHost,
		ServerIDKey:   domain.DefaultSpeedTestNetServerID,
	}

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
}

func TestGetSpeedTestStringValuesWithNamedServerCustomServer(t *testing.T) {
	serverHost := "SpeedTestHost"
	namedServerUID := "nst21"

	namedServerFixtures := getCustomNamedServerFixtures(namedServerUID, serverHost)
	db.LoadNamedServerFixtures(namedServerFixtures, t)

	task := domain.Task{}
	task.NamedServer = namedServerFixtures[0]

	results, err := getSpeedTestStringValues(task)
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: serverHost,
		ServerIDKey:   namedServerUID,
	}

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
}

func TestGetSpeedTestStringValuesWithSpeedTestServer(t *testing.T) {
	serverID := "111"
	serverHost := "SpeedTestHost"
	country := domain.Country{Code: "US", Name: "United States"}

	sTNetServerListFixtures := []domain.STNetServerList{
		{
			ID:      domain.DataTypeSTNetServerList + "-" + country.Code,
			Country: country,
			Servers: []domain.SpeedTestNetServer{
				domain.SpeedTestNetServer{Host: serverHost, ServerID: serverID},
			},
		},
	}

	db.LoadSTNetServerListFixtures(sTNetServerListFixtures, t)

	namedServerUID := "nst22"

	namedServerFixtures := getCustomNamedServerFixtures(namedServerUID, serverHost)
	namedServerFixtures[0].ServerType = domain.ServerTypeSpeedTestNet
	namedServerFixtures[0].Country = country
	namedServerFixtures[0].SpeedTestNetServerID = serverID

	db.LoadNamedServerFixtures(namedServerFixtures, t)

	task := domain.Task{}
	task.NamedServer = namedServerFixtures[0]

	results, err := getSpeedTestStringValues(task)
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: serverHost,
		ServerIDKey:   serverID,
	}

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
}

func TestUpdateTaskSpeedTestWithoutNamedServer(t *testing.T) {
	task := domain.Task{}
	task.NamedServer = domain.NamedServer{}

	resultsTask, err := updateTaskSpeedTest(task)

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: domain.DefaultSpeedTestNetServerHost,
		ServerIDKey:   domain.DefaultSpeedTestNetServerID,
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
	resultsInts := resultsTask.Data.IntValues
	expectedInts := map[string]int{TimeOutKey: DefaultSpeedTestTimeoutInSeconds}

	if !areIntMapsEqual(expectedInts, resultsInts) {
		t.Errorf("Bad IntValues.\nExpected: %v.\n But got: %v", expectedInts, resultsInts)
	}

	resultsFloats := resultsTask.Data.FloatValues
	expectedFloats := map[string]float64{MaxSecondsKey: DefaultSpeedTestMaxSeconds}

	if !areFloatMapsEqual(expectedFloats, resultsFloats) {
		t.Errorf("Bad FloatValues.\nExpected: %v.\n But got: %v", expectedFloats, resultsFloats)
	}

	resultsIntSlices := resultsTask.Data.IntSlices
	expectedIntSlices := map[string][]int{
		DownloadSizesKey: GetDefaultSpeedTestDownloadSizes(),
		UploadSizesKey:   GetDefaultSpeedTestUploadSizes(),
	}

	if !areIntSliceMapsEqual(expectedIntSlices, resultsIntSlices) {
		t.Errorf("Bad FloatValues.\nExpected: %v.\n But got: %v", expectedIntSlices, resultsIntSlices)
	}
}

func TestUpdateTaskSpeedTestWithNamedServerCustomServer(t *testing.T) {
	serverHost := "SpeedTestHost"
	namedServerUID := "nst23"

	namedServerFixtures := getCustomNamedServerFixtures(namedServerUID, serverHost)
	db.LoadNamedServerFixtures(namedServerFixtures, t)

	task := domain.Task{}
	task.NamedServer = namedServerFixtures[0]

	resultsTask, err := updateTaskSpeedTest(task)

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: serverHost,
		ServerIDKey:   namedServerUID,
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
	resultsInts := resultsTask.Data.IntValues
	expectedInts := map[string]int{TimeOutKey: DefaultSpeedTestTimeoutInSeconds}

	if !areIntMapsEqual(expectedInts, resultsInts) {
		t.Errorf("Bad IntValues.\nExpected: %v.\n But got: %v", expectedInts, resultsInts)
	}

	resultsFloats := resultsTask.Data.FloatValues
	expectedFloats := map[string]float64{MaxSecondsKey: DefaultSpeedTestMaxSeconds}

	if !areFloatMapsEqual(expectedFloats, resultsFloats) {
		t.Errorf("Bad FloatValues.\nExpected: %v.\n But got: %v", expectedFloats, resultsFloats)
	}

	resultsIntSlices := resultsTask.Data.IntSlices
	expectedIntSlices := map[string][]int{
		DownloadSizesKey: GetDefaultSpeedTestDownloadSizes(),
		UploadSizesKey:   GetDefaultSpeedTestUploadSizes(),
	}

	if !areIntSliceMapsEqual(expectedIntSlices, resultsIntSlices) {
		t.Errorf("Bad FloatValues.\nExpected: %v.\n But got: %v", expectedIntSlices, resultsIntSlices)
	}
}

func TestUpdateTaskSpeedTestWithSpeedTestNetServer(t *testing.T) {
	serverID := "111"
	serverHost := "SpeedTestHost"
	country := domain.Country{Code: "US", Name: "United States"}

	sTNetServerListFixtures := []domain.STNetServerList{
		{
			ID:      domain.DataTypeSTNetServerList + "-" + country.Code,
			Country: country,
			Servers: []domain.SpeedTestNetServer{
				domain.SpeedTestNetServer{Host: serverHost, ServerID: serverID},
			},
		},
	}

	db.LoadSTNetServerListFixtures(sTNetServerListFixtures, t)

	namedServerUID := "nst23"

	namedServerFixtures := getCustomNamedServerFixtures(namedServerUID, serverHost)
	namedServerFixtures[0].ServerType = domain.ServerTypeSpeedTestNet
	namedServerFixtures[0].Country = country
	namedServerFixtures[0].SpeedTestNetServerID = serverID

	db.LoadNamedServerFixtures(namedServerFixtures, t)

	task := domain.Task{}
	task.NamedServer = namedServerFixtures[0]

	resultsTask, err := updateTaskSpeedTest(task)

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: serverHost,
		ServerIDKey:   serverID,
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}
	resultsInts := resultsTask.Data.IntValues
	expectedInts := map[string]int{TimeOutKey: DefaultSpeedTestTimeoutInSeconds}

	if !areIntMapsEqual(expectedInts, resultsInts) {
		t.Errorf("Bad IntValues.\nExpected: %v.\n But got: %v", expectedInts, resultsInts)
	}

	resultsFloats := resultsTask.Data.FloatValues
	expectedFloats := map[string]float64{MaxSecondsKey: DefaultSpeedTestMaxSeconds}

	if !areFloatMapsEqual(expectedFloats, resultsFloats) {
		t.Errorf("Bad FloatValues.\nExpected: %v.\n But got: %v", expectedFloats, resultsFloats)
	}

	resultsIntSlices := resultsTask.Data.IntSlices
	expectedIntSlices := map[string][]int{
		DownloadSizesKey: GetDefaultSpeedTestDownloadSizes(),
		UploadSizesKey:   GetDefaultSpeedTestUploadSizes(),
	}

	if !areIntSliceMapsEqual(expectedIntSlices, resultsIntSlices) {
		t.Errorf("Bad FloatValues.\nExpected: %v.\n But got: %v", expectedIntSlices, resultsIntSlices)
	}
}

func TestUpdateNodeTasksWithPingWithoutNamedServer(t *testing.T) {
	task := domain.Task{}
	task.Type = domain.TaskTypePing
	task.NamedServer = domain.NamedServer{}
	node := domain.Node{}

	node.Tasks = []domain.Task{task}

	resultsNode, err := updateNodeTasks(node)

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsNode.Tasks[0].Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: domain.DefaultPingServerHost,
		ServerIDKey:   domain.DefaultPingServerID,
	}

	if !areStringMapsEqual(expected, results) {
		t.Errorf("Bad StringValues.\nExpected: %v.\n But got: %v", expected, results)
	}

	resultsInts := resultsNode.Tasks[0].Data.IntValues
	expectedInts := map[string]int{TimeOutKey: DefaultPingTimeoutInSeconds}

	if !areIntMapsEqual(expectedInts, resultsInts) {
		t.Errorf("Bad IntValues.\nExpected: %v.\n But got: %v", expectedInts, resultsInts)
	}
}

func TestNodeAuthorization(t *testing.T) {
	db.FlushTables(t)
	db.LoadTagFixtures(getTagFixtures(), t)

	users := getUserFixtures()
	db.LoadUserFixtures(users, t)

	nodeFixtures := getNodeFixtures()
	db.LoadNodeFixtures(nodeFixtures, t)

	// Superadmin user
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/aa:aa:aa:aa:aa:aa",
		PathParameters: map[string]string{
			"macAddr": "aa:aa:aa:aa:aa:aa",
		},
		Headers: map[string]string{
			domain.UserReqHeaderID: "super_admin",
		},
	}

	response, err := viewNode(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
	}

	// Authorized admin user
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/aa:aa:aa:aa:aa:aa",
		PathParameters: map[string]string{
			"macAddr": "aa:aa:aa:aa:aa:aa",
		},
		Headers: map[string]string{
			domain.UserReqHeaderID: "pass_test",
		},
	}

	response, err = viewNode(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
	}

	// Unauthorized admin user should be rejected
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/report/node/aa:aa:aa:aa:aa:aa",
		PathParameters: map[string]string{
			"macAddr": "aa:aa:aa:aa:aa:aa",
		},
		Headers: map[string]string{
			domain.UserReqHeaderID: "fail_test",
		},
	}

	response, err = viewNode(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 403 {
		t.Error("Wrong status code returned, expected 403, got", response.StatusCode, response.Body)
	}

}

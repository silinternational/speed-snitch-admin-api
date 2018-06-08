package main

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"reflect"
	"strings"
	"testing"
)

type DBClient struct{}

const TestHostForSpeedTestNet = "SpeedTestNetFixtureHost"
const TestServerIDForSpeedTestNet = "111"

// For test fixtures, the value param is going to dictate the attribute values.
//   The value should have this format ...
//     UID|ServerType|ServerHost|SpeedTestNetServerID
func (d DBClient) GetItem(tableAlias, dataType, value string, itemObj interface{}) error {
	desiredAttributes := map[string]string{}
	fixtureValues := strings.Split(value, "|")
	desiredAttributes["UID"] = fixtureValues[0]
	desiredAttributes["ServerType"] = fixtureValues[1]
	desiredAttributes["ServerHost"] = fixtureValues[2]
	desiredAttributes["SpeedTestNetServerID"] = fixtureValues[3]

	stype := reflect.ValueOf(itemObj).Elem()
	for fieldName, value := range desiredAttributes {
		field := stype.FieldByName(fieldName)
		if field.IsValid() {
			field.SetString(value)
		} else {
			return fmt.Errorf("Can't set value on attribute: %s", fieldName)
		}
	}
	return nil
}

// Minimal test fixture
func (d DBClient) GetSpeedTestNetServerFromNamedServer(namedServer domain.NamedServer) (domain.SpeedTestNetServer, error) {
	stnServer := domain.SpeedTestNetServer{
		ServerID: TestServerIDForSpeedTestNet,
		Host:     TestHostForSpeedTestNet,
	}
	return stnServer, nil
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
	results, err := getPingStringValues(task, DBClient{})
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
	task := domain.Task{}
	uid := "NSTest1"
	serverType := domain.ServerTypeCustom
	serverHost := "PingTestHost"
	task.NamedServer = domain.NamedServer{ID: uid + "|" + serverType + "|" + serverHost + "|"}

	results, err := getPingStringValues(task, DBClient{})
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: serverHost,
		ServerIDKey:   uid,
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

	resultsTask, err := updateTaskPing(task, DBClient{})

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
	task := domain.Task{}
	uid := "NSTest1"
	serverType := domain.ServerTypeCustom
	serverHost := "SpeedTestHost"
	task.NamedServer = domain.NamedServer{ID: uid + "|" + serverType + "|" + serverHost + "|"}

	resultsTask, err := updateTaskPing(task, DBClient{})

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigLatencyTest,
		ServerHostKey: serverHost,
		ServerIDKey:   uid,
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

	results, err := getSpeedTestStringValues(task, DBClient{})
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
	task := domain.Task{}
	uid := "NSTest2"
	serverType := domain.ServerTypeCustom
	serverHost := "SpeedTestHost"
	task.NamedServer = domain.NamedServer{ID: uid + "|" + serverType + "|" + serverHost + "|"}

	results, err := getSpeedTestStringValues(task, DBClient{})
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: serverHost,
		ServerIDKey:   uid,
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
	task := domain.Task{}
	uid := "NSTest3"
	serverType := domain.ServerTypeSpeedTestNet
	serverHost := "SpeedTestNetHost"
	speedTestNetServerID := TestServerIDForSpeedTestNet
	task.NamedServer = domain.NamedServer{ID: uid + "|" + serverType + "|" + serverHost + "|" + speedTestNetServerID}

	results, err := getSpeedTestStringValues(task, DBClient{})
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: TestHostForSpeedTestNet,
		ServerIDKey:   TestServerIDForSpeedTestNet,
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

	resultsTask, err := updateTaskSpeedTest(task, DBClient{})

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
	task := domain.Task{}
	uid := "NSTest2"
	serverType := domain.ServerTypeCustom
	serverHost := "SpeedTestHost"
	task.NamedServer = domain.NamedServer{ID: uid + "|" + serverType + "|" + serverHost + "|"}

	resultsTask, err := updateTaskSpeedTest(task, DBClient{})

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: serverHost,
		ServerIDKey:   uid,
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
	task := domain.Task{}
	uid := "NSTest3"
	serverType := domain.ServerTypeSpeedTestNet
	serverHost := "SpeedTestHost"
	task.NamedServer = domain.NamedServer{ID: uid + "|" + serverType + "|" + serverHost + "|"}

	resultsTask, err := updateTaskSpeedTest(task, DBClient{})

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	results := resultsTask.Data.StringValues
	expected := map[string]string{
		TestTypeKey:   domain.TestConfigSpeedTest,
		ServerHostKey: TestHostForSpeedTestNet,
		ServerIDKey:   TestServerIDForSpeedTestNet,
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

	resultsNode, err := updateNodeTasks(node, DBClient{})

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

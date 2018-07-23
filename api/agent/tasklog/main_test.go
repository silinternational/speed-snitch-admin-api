package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"testing"
	"time"
)

func TestHandlerSpeedTest(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		MacAddr:     "aa:aa:aa:aa:aa:aa",
		IPAddress:   "123.123.123.123",
		Location:    "Charlotte, NC, Unitied States",
		Coordinates: "23,23",
		RunningVersion: domain.Version{
			Model: gorm.Model{
				ID: 1,
			},
			Number: "1.0.0",
		},
	}
	db.PutItem(&node1)

	speedTestNetServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 1,
		},
		CountryCode: "US",
		Country:     "United States",
		ServerID:    "1234",
		Name:        "test-server",
		Host:        "test.com:8080",
		Lat:         "123",
		Lon:         "123",
	}
	db.PutItem(&speedTestNetServer)

	namedServer := domain.NamedServer{
		Model: gorm.Model{
			ID: 1,
		},
		ServerType:           domain.DataTypeSpeedTestNetServer,
		SpeedTestNetServerID: speedTestNetServer.ID,
		Name:                 "test example",
		ServerCountry:        speedTestNetServer.CountryCode,
		ServerHost:           speedTestNetServer.Host,
	}
	db.PutItem(&namedServer)

	logsToSend := []domain.TaskLogSpeedTest{
		{
			Timestamp:     1531246102,
			NamedServerID: namedServer.ID,
			Upload:        10,
			Download:      10,
		},
		{
			Timestamp:     1531246103,
			NamedServerID: namedServer.ID,
			Upload:        20,
			Download:      20,
		},
		{
			Timestamp:     1531246104,
			NamedServerID: namedServer.ID,
			Upload:        30,
			Download:      30,
		},
	}

	for _, i := range logsToSend {
		js, err := json.Marshal(i)
		if err != nil {
			t.Error("Unable to marshal log fixture to json, err: ", err.Error())
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       fmt.Sprintf("/log/%s/%s", node1.MacAddr, domain.TaskTypeSpeedTest),
			PathParameters: map[string]string{
				"macAddr":   node1.MacAddr,
				"entryType": domain.TaskTypeSpeedTest,
			},
			Body: string(js),
		}

		resp, err := Handler(req)
		if err != nil {
			t.Errorf("Got error trying to submit log for timestamp: %v, err: %s", i.Timestamp, err.Error())
		}

		if resp.StatusCode != 204 {
			t.Errorf("Go incorrect status code submitting log, expected 204, got %v. body: %s", resp.StatusCode, resp.Body)
		}
	}

	var taskLogEntry domain.TaskLogSpeedTest
	err := db.GetTaskLogForRange(&taskLogEntry, node1.ID, logsToSend[0].Timestamp, logsToSend[0].Timestamp)
	if err != nil {
		t.Error("Unable to retrieve task log for first entry, err: ", err.Error())
	}

	if taskLogEntry.ServerCountry != speedTestNetServer.CountryCode {
		t.Errorf("Task log entry was not created properly, wrong country code, expected %s, got %s", speedTestNetServer.CountryCode, taskLogEntry.ServerCountry)
	}
	if taskLogEntry.Upload != logsToSend[0].Upload {
		t.Errorf("Task log entry does not have correct upload value, expected %v, got %v", logsToSend[0].Upload, taskLogEntry.Upload)
	}
}

func TestHandlerPing(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		MacAddr:     "aa:aa:aa:aa:aa:aa",
		IPAddress:   "123.123.123.123",
		Location:    "Charlotte, NC, Unitied States",
		Coordinates: "23,23",
		RunningVersion: domain.Version{
			Model: gorm.Model{
				ID: 1,
			},
			Number: "1.0.0",
		},
	}
	db.PutItem(&node1)

	namedServer := domain.NamedServer{
		Model: gorm.Model{
			ID: 1,
		},
		ServerType:    domain.ServerTypeCustom,
		ServerHost:    "google.com",
		ServerCountry: "US",
		Name:          "test example",
	}
	db.PutItem(&namedServer)

	logsToSend := []domain.TaskLogPingTest{
		{
			Timestamp:         1531246102,
			NamedServerID:     namedServer.ID,
			Latency:           0.5,
			PacketLossPercent: 1,
		},
		{
			Timestamp:         1531246103,
			NamedServerID:     namedServer.ID,
			Latency:           1,
			PacketLossPercent: 2,
		},
		{
			Timestamp:         1531246104,
			NamedServerID:     namedServer.ID,
			Latency:           3,
			PacketLossPercent: 0,
		},
	}

	for _, i := range logsToSend {
		js, err := json.Marshal(i)
		if err != nil {
			t.Error("Unable to marshal log fixture to json, err: ", err.Error())
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       fmt.Sprintf("/log/%s/%s", node1.MacAddr, domain.TaskTypePing),
			PathParameters: map[string]string{
				"macAddr":   node1.MacAddr,
				"entryType": domain.TaskTypePing,
			},
			Body: string(js),
		}

		resp, err := Handler(req)
		if err != nil {
			t.Errorf("Got error trying to submit log for timestamp: %v, err: %s", i.Timestamp, err.Error())
		}

		if resp.StatusCode != 204 {
			t.Errorf("Go incorrect status code submitting log, expected 204, got %v. body: %s", resp.StatusCode, resp.Body)
		}
	}

	var taskLogEntry domain.TaskLogPingTest
	err := db.GetTaskLogForRange(&taskLogEntry, node1.ID, logsToSend[0].Timestamp, logsToSend[0].Timestamp)
	if err != nil {
		t.Error("Unable to retrieve task log for first entry, err: ", err.Error())
	}

	if taskLogEntry.ServerCountry != namedServer.ServerCountry {
		t.Errorf("Task log entry was not created properly, wrong country code, expected %s, got %s", namedServer.ServerCountry, taskLogEntry.ServerCountry)
	}
	if taskLogEntry.Latency != logsToSend[0].Latency {
		t.Errorf("Task log entry does not have correct upload value, expected %v, got %v", logsToSend[0].Latency, taskLogEntry.Latency)
	}
}

func TestHandlerDowntime(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		MacAddr:     "aa:aa:aa:aa:aa:aa",
		IPAddress:   "123.123.123.123",
		Location:    "Charlotte, NC, Unitied States",
		Coordinates: "23,23",
		RunningVersion: domain.Version{
			Model: gorm.Model{
				ID: 1,
			},
			Number: "1.0.0",
		},
	}
	db.PutItem(&node1)

	logsToSend := []domain.TaskLogNetworkDowntime{
		{
			Timestamp:       1531246102,
			DowntimeStart:   time.Now().UTC().Format(time.RFC3339),
			DowntimeSeconds: 500,
		},
		{
			Timestamp:       1531246103,
			DowntimeStart:   time.Now().UTC().Format(time.RFC3339),
			DowntimeSeconds: 1000,
		},
		{
			Timestamp:       1531246104,
			DowntimeStart:   time.Now().UTC().Format(time.RFC3339),
			DowntimeSeconds: 5000,
		},
	}

	for _, i := range logsToSend {
		js, err := json.Marshal(i)
		if err != nil {
			t.Error("Unable to marshal log fixture to json, err: ", err.Error())
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       fmt.Sprintf("/log/%s/%s", node1.MacAddr, domain.LogTypeDowntime),
			PathParameters: map[string]string{
				"macAddr":   node1.MacAddr,
				"entryType": domain.LogTypeDowntime,
			},
			Body: string(js),
		}

		resp, err := Handler(req)
		if err != nil {
			t.Errorf("Got error trying to submit log for timestamp: %v, err: %s", i.Timestamp, err.Error())
		}

		if resp.StatusCode != 204 {
			t.Errorf("Go incorrect status code submitting log, expected 204, got %v. body: %s", resp.StatusCode, resp.Body)
		}
	}

	var taskLogEntry domain.TaskLogNetworkDowntime
	err := db.GetTaskLogForRange(&taskLogEntry, node1.ID, logsToSend[0].Timestamp, logsToSend[0].Timestamp)
	if err != nil {
		t.Error("Unable to retrieve task log for first entry, err: ", err.Error())
	}

	if taskLogEntry.DowntimeStart != logsToSend[0].DowntimeStart {
		t.Errorf("Task log entry was not created properly, wrong downtime start, expected %s, got %s", logsToSend[0].DowntimeStart, taskLogEntry.DowntimeStart)
	}
	if taskLogEntry.DowntimeSeconds != logsToSend[0].DowntimeSeconds {
		t.Errorf("Task log entry does not have correct downtime seconds value, expected %v, got %v", logsToSend[0].DowntimeSeconds, taskLogEntry.DowntimeSeconds)
	}
}

func TestHandlerRestart(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		MacAddr:     "aa:aa:aa:aa:aa:aa",
		IPAddress:   "123.123.123.123",
		Location:    "Charlotte, NC, Unitied States",
		Coordinates: "23,23",
		RunningVersion: domain.Version{
			Model: gorm.Model{
				ID: 1,
			},
			Number: "1.0.0",
		},
	}
	db.PutItem(&node1)

	logsToSend := []domain.TaskLogRestart{
		{
			Timestamp: 1531246102,
		},
		{
			Timestamp: 1531246103,
		},
		{
			Timestamp: 1531246104,
		},
	}

	for _, i := range logsToSend {
		js, err := json.Marshal(i)
		if err != nil {
			t.Error("Unable to marshal log fixture to json, err: ", err.Error())
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       fmt.Sprintf("/log/%s/%s", node1.MacAddr, domain.LogTypeRestart),
			PathParameters: map[string]string{
				"macAddr":   node1.MacAddr,
				"entryType": domain.LogTypeRestart,
			},
			Body: string(js),
		}

		resp, err := Handler(req)
		if err != nil {
			t.Errorf("Got error trying to submit log for timestamp: %v, err: %s", i.Timestamp, err.Error())
		}

		if resp.StatusCode != 204 {
			t.Errorf("Go incorrect status code submitting log, expected 204, got %v. body: %s", resp.StatusCode, resp.Body)
		}
	}

	var taskLogEntry []domain.TaskLogRestart
	err := db.GetTaskLogForRange(&taskLogEntry, node1.ID, logsToSend[0].Timestamp, logsToSend[2].Timestamp)
	if err != nil {
		t.Error("Unable to retrieve task log for first entry, err: ", err.Error())
	}

	if len(taskLogEntry) != 3 {
		t.Errorf("Did not get back expected number of restart logs, expected %v, got %v", 3, len(taskLogEntry))
	}
}

func TestHandlerError(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		MacAddr:     "aa:aa:aa:aa:aa:aa",
		IPAddress:   "123.123.123.123",
		Location:    "Charlotte, NC, Unitied States",
		Coordinates: "23,23",
		RunningVersion: domain.Version{
			Model: gorm.Model{
				ID: 1,
			},
			Number: "1.0.0",
		},
	}
	db.PutItem(&node1)

	speedTestNetServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 1,
		},
		CountryCode: "US",
		Country:     "United States",
		ServerID:    "1234",
		Name:        "test-server",
		Host:        "test.com:8080",
		Lat:         "123",
		Lon:         "123",
	}
	db.PutItem(&speedTestNetServer)

	namedServer := domain.NamedServer{
		Model: gorm.Model{
			ID: 1,
		},
		ServerType:           domain.DataTypeSpeedTestNetServer,
		SpeedTestNetServerID: speedTestNetServer.ID,
		Name:                 "test example",
		ServerCountry:        speedTestNetServer.CountryCode,
		ServerHost:           speedTestNetServer.Host,
	}
	db.PutItem(&namedServer)

	logsToSend := []domain.TaskLogError{
		{
			Timestamp:     1531246102,
			NamedServerID: namedServer.ID,
			ErrorCode:     "abc123",
			ErrorMessage:  "error message",
		},
		{
			Timestamp:     1531246103,
			NamedServerID: namedServer.ID,
			ErrorCode:     "abc123",
			ErrorMessage:  "error message",
		},
		{
			Timestamp:     1531246104,
			NamedServerID: namedServer.ID,
			ErrorCode:     "abc123",
			ErrorMessage:  "error message",
		},
	}

	for _, i := range logsToSend {
		js, err := json.Marshal(i)
		if err != nil {
			t.Error("Unable to marshal log fixture to json, err: ", err.Error())
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       fmt.Sprintf("/log/%s/%s", node1.MacAddr, domain.LogTypeError),
			PathParameters: map[string]string{
				"macAddr":   node1.MacAddr,
				"entryType": domain.LogTypeError,
			},
			Body: string(js),
		}

		resp, err := Handler(req)
		if err != nil {
			t.Errorf("Got error trying to submit log for timestamp: %v, err: %s", i.Timestamp, err.Error())
		}

		if resp.StatusCode != 204 {
			t.Errorf("Go incorrect status code submitting log, expected 204, got %v. body: %s", resp.StatusCode, resp.Body)
		}
	}

	var taskLogEntry domain.TaskLogError
	err := db.GetTaskLogForRange(&taskLogEntry, node1.ID, logsToSend[0].Timestamp, logsToSend[0].Timestamp)
	if err != nil {
		t.Error("Unable to retrieve task log for first entry, err: ", err.Error())
	}

	if taskLogEntry.ServerCountry != speedTestNetServer.CountryCode {
		t.Errorf("Task log entry was not created properly, wrong country code, expected %s, got %s", speedTestNetServer.CountryCode, taskLogEntry.ServerCountry)
	}
	if taskLogEntry.ErrorCode != logsToSend[0].ErrorCode {
		t.Errorf("Task log entry does not have correct upload value, expected %v, got %v", logsToSend[0].ErrorCode, taskLogEntry.ErrorCode)
	}
}

package db

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-agent"
	"github.com/silinternational/speed-snitch-agent/lib/speedtestnet"
	"testing"
)

var testTasks = map[string]agent.Task{
	"111Ping": {
		Type:     agent.TypePing,
		Schedule: "*/10 * * * *",
		Data: agent.TaskData{
			StringValues: map[string]string{
				speedtestnet.CFG_TEST_TYPE:   speedtestnet.CFG_TYPE_LATENCY,
				speedtestnet.CFG_SERVER_HOST: "host1.net:8080",
			},
			IntValues: map[string]int{
				speedtestnet.CFG_SERVER_ID: 111,
			},
		},
	},
	"111SpeedTest": {
		Type:     agent.TypeSpeedTest,
		Schedule: "* 2 * * *",
		Data: agent.TaskData{
			StringValues: map[string]string{
				speedtestnet.CFG_TEST_TYPE:   speedtestnet.CFG_TYPE_ALL,
				speedtestnet.CFG_SERVER_HOST: "host1.net:8080",
			},
			IntValues: map[string]int{
				speedtestnet.CFG_SERVER_ID: 111,
			},
		},
	},
	"222Ping": {
		Type:     agent.TypePing,
		Schedule: "*/30 * * * *",
		Data: agent.TaskData{
			StringValues: map[string]string{
				speedtestnet.CFG_TEST_TYPE:   speedtestnet.CFG_TYPE_LATENCY,
				speedtestnet.CFG_SERVER_HOST: "host2.net:8080",
			},
			IntValues: map[string]int{
				speedtestnet.CFG_SERVER_ID: 222,
			},
		},
	},
}

func TestGetServerDataFromNode(t *testing.T) {

	node := domain.Node{
		"11:22:33:44:55:66",
		"linux",
		"amd",
		"0.0.1",
		"0.0.1",
		"10:11:12",
		"1/3/2018",
		"1/1/2018",
		"Kenya, , Nairobi",
		"1.2.3.4",
		[]agent.Task{
			testTasks["111Ping"],
			testTasks["111SpeedTest"],
			testTasks["222Ping"],
		},
		[]domain.Contact{
			{
				Name:  "Andy Admin",
				Email: "andy_admin@some.org",
				Phone: "100-123-4567",
			},
		},
		[]domain.Tag{domain.Tag{Name: "Field", Description: "Field location"}},
		"John Doe",
	}

	servers, err := GetServerDataFromNode(node)

	if err != nil {
		t.Errorf("Did not expect to get an error, but got ...\n\t%s", err.Error())
		return
	}

	results := len(servers)
	expected := 2

	if expected != results {
		t.Errorf("Wrong number of servers. Expected: %d. But got: %d", expected, results)
		return
	}

	results = servers[0].ID
	expected = 111

	if expected != results {
		t.Errorf("Wrong server ID. Expected: %d. But got: %d", expected, results)
		return
	}

	resultsHost := servers[0].Host
	expectedHost := "host1.net:8080"

	if expectedHost != resultsHost {
		t.Errorf("Wrong server Host. Expected: %s. But got: %s", expectedHost, resultsHost)
		return
	}
}

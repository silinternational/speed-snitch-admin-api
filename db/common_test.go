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
		Schedule: "*/11 * * * *",
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
		Schedule: "* 1 * * *",
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
		Schedule: "*/22 * * * *",
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
	"222SpeedTest": {
		Type:     agent.TypeSpeedTest,
		Schedule: "* 2 * * *",
		Data: agent.TaskData{
			StringValues: map[string]string{
				speedtestnet.CFG_TEST_TYPE:   speedtestnet.CFG_TYPE_ALL,
				speedtestnet.CFG_SERVER_HOST: "host2.net:8080",
			},
			IntValues: map[string]int{
				speedtestnet.CFG_SERVER_ID: 222,
			},
		},
	},
}

var testNodes = map[string]domain.Node{
	"11Kenya": {
		"1111",
		"11:11:11:11:11:11",
		"linux",
		"amd",
		"0.0.1",
		"0.0.1",
		1111,
		"1/3/2018",
		"1/1/2018",
		"Kenya, , Nairobi",
		"1Lat 1Lon",
		"1.1.1.1",
		"11.11.11.11",
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
	},
	"22Chad": {
		"2222",
		"22:22:22:22:22:22",
		"linux",
		"amd",
		"0.0.2",
		"0.0.2",
		2222,
		"1/3/2018",
		"1/1/2018",
		"Chad, , N'Djamena",
		"2Lat 2Lon",
		"2.2.2.2",
		"22.22.22.22",
		[]agent.Task{
			testTasks["222Ping"],
			testTasks["222SpeedTest"],
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
	},
}

func TestGetServerDataFromNode(t *testing.T) {
	node := testNodes["11Kenya"]

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

func TestGetNodesForServers(t *testing.T) {
	nodes := []domain.Node{testNodes["11Kenya"], testNodes["22Chad"]}

	nodesForServers, err := GetNodesForServers(nodes)
	if err != nil {
		t.Errorf("Did not expect to get an error, but got ...\n\t%s", err.Error())
		return
	}

	results := len(nodesForServers)
	expected := 2
	if expected != results {
		t.Errorf("Wrong number of servers. Expected: %d. But got: %d", expected, results)
		return
	}

	id := 111
	results = len(nodesForServers[id].Nodes)
	expected = 1
	if expected != results {
		t.Errorf("Wrong number of nodes for server %d. Expected: %d. But got: %d", id, expected, results)
		return
	}

	id = 222
	results = len(nodesForServers[id].Nodes)
	expected = 2
	if expected != results {
		t.Errorf("Wrong number of nodes for server %d. Expected: %d. But got: %d", id, expected, results)
		return
	}

	expectedIP := testNodes["11Kenya"].IPAddress
	resultsIP := nodesForServers[id].Nodes[0].IPAddress
	if expectedIP != resultsIP {
		t.Errorf("Wrong IP address for 1st node for server %d. Expected: %s. But got: %s", id, expectedIP, resultsIP)
		return
	}

}

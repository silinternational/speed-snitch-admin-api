package db

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-agent"
	"github.com/silinternational/speed-snitch-agent/lib/speedtestnet"
	"strings"
	"testing"
)

var testTasks = map[string]domain.Task{
	"111Ping": {
		Type:     agent.TypePing,
		Schedule: "*/11 * * * *",
		NamedServer: domain.NamedServer{
			ID:   "namedserver-000",
			UID:  "000",
			Name: "Outdated NamedServer",
		},
		Data: domain.TaskData{
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
		NamedServer: domain.NamedServer{
			ID:   "namedserver-999",
			UID:  "999",
			Name: "Deleted NamedServer",
		},
		Data: domain.TaskData{
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
		Data: domain.TaskData{
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
		Data: domain.TaskData{
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
		ID:                "node-11:11:11:11:11:11",
		MacAddr:           "11:11:11:11:11:11",
		OS:                "linux",
		Arch:              "amd",
		RunningVersion:    "0.0.1",
		ConfiguredVersion: "0.0.1",
		Uptime:            1111,
		LastSeen:          "1/3/2018",
		FirstSeen:         "1/1/2018",
		Location:          "Kenya, , Nairobi",
		Coordinates:       "1Lat 1Lon",
		IPAddress:         "1.1.1.1",
		Tasks: []domain.Task{
			testTasks["111Ping"],
			testTasks["111SpeedTest"],
			testTasks["222Ping"],
		},
		Contacts: []domain.Contact{
			{
				Name:  "Andy Admin",
				Email: "andy_admin@some.org",
				Phone: "100-123-4567",
			},
		},
		Tags: []domain.Tag{
			{ID: "tag-000", UID: "000", Name: "Eastern Africa"},
			{ID: "tag-111", UID: "111", Name: "Anglophone Africa"},
		},
		ConfiguredBy: "John Doe",
		Nickname:     "Nairobi RaspberryPi",
		Notes:        "",
	},
	"22Chad": {
		ID:                "node-22:22:22:22:22:22",
		MacAddr:           "22:22:22:22:22:22",
		OS:                "linux",
		Arch:              "amd",
		RunningVersion:    "0.0.2",
		ConfiguredVersion: "0.0.2",
		Uptime:            2222,
		LastSeen:          "1/3/2018",
		FirstSeen:         "1/1/2018",
		Location:          "Chad, , N'Djamena",
		Coordinates:       "2Lat 2Lon",
		IPAddress:         "2.2.2.2",
		Tasks: []domain.Task{
			testTasks["222Ping"],
			testTasks["222SpeedTest"],
		},
		Contacts: []domain.Contact{
			{
				Name:  "Andy Admin",
				Email: "andy_admin@some.org",
				Phone: "100-123-4567",
			},
		},
		Tags:         []domain.Tag{{ID: "tag-222", UID: "222"}, {ID: "tag-333", UID: "333"}},
		ConfiguredBy: "John Doe",
		Nickname:     "Chad Windoze server",
		Notes:        "",
	},
}

var tagFixtures = []domain.Tag{
	{ID: "tag-000", UID: "000", Name: "Test Tag 000"},
	{ID: "tag-111", UID: "111", Name: "Test Tag 111"},
	{ID: "tag-222", UID: "222", Name: "Test Tag 222"},
	{ID: "tag-333", UID: "333", Name: "Test Tag 333"},
}

var namedServerFixtures = []domain.NamedServer{
	{ID: "namedserver-000", UID: "000", Name: "New Name"},
}

func TestUpdateTags(t *testing.T) {
	FlushTables(t)
	LoadTagFixtures(tagFixtures, t)

	oldTags := []domain.Tag{
		{ID: "tag-000", UID: "000", Name: "Bad Name 000"},
		tagFixtures[1],
		{ID: "tag-999", UID: "999", Name: "Doesn't Exist"},
	}

	results, err := updateTags(oldTags)
	if err != nil {
		t.Errorf("Unexpected Error. \n%s", err.Error())
		return
	}

	expected := []domain.Tag{tagFixtures[0], tagFixtures[1]}

	areTagsEqual(expected, results, t)
}

// Make sure the updated task's NamedServer gets the new information for the matching speedtestnet server
func TestGetUpdatedTasks(t *testing.T) {
	serverID := "111"
	newServerHost := "new.host.com" // This should be a change
	country := domain.Country{Code: "US", Name: "United States"}

	sTNetServerListFixtures := []domain.STNetServerList{
		{
			ID:      domain.DataTypeSTNetServerList + "-" + country.Code,
			Country: country,
			Servers: []domain.SpeedTestNetServer{
				domain.SpeedTestNetServer{Host: newServerHost, ServerID: serverID},
			},
		},
	}

	LoadSTNetServerListFixtures(sTNetServerListFixtures, t)

	namedServerFixtures := []domain.NamedServer{
		{
			ID:                   "namedserver-000",
			UID:                  "000",
			Name:                 "New Name",
			Country:              country,
			ServerType:           domain.ServerTypeSpeedTestNet,
			ServerHost:           "outdated.host.com", // This should get changed
			SpeedTestNetServerID: serverID,
		},
	}

	LoadNamedServerFixtures(namedServerFixtures, t)

	tasks := []domain.Task{
		{
			Type:     "reboot",
			Schedule: "*/5 * * * *",
		},
		testTasks["111Ping"],
		testTasks["222SpeedTest"],
	}

	results, err := GetUpdatedTasks(tasks)

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	expected := []domain.Task{
		{
			Type:     "reboot",
			Schedule: "*/5 * * * *",
		},
		{
			Type:     "ping",
			Schedule: "*/11 * * * *",
			NamedServer: domain.NamedServer{
				ID:                   "namedserver-000",
				UID:                  "000",
				Name:                 "New Name",
				Country:              country,
				ServerType:           domain.ServerTypeSpeedTestNet,
				ServerHost:           newServerHost, // This should have gotten changed
				SpeedTestNetServerID: serverID,
			},
			ServerHost:           newServerHost, // This should have gotten changed
			SpeedTestNetServerID: serverID,
		},
		{
			Type:        "speedTest",
			Schedule:    "* 2 * * *",
			NamedServer: domain.NamedServer{},
		},
	}

	gotError := false

	if len(results) != len(expected) {
		gotError = true
	}

	if !gotError {
		for index, nextExpected := range expected {
			if results[index].Type != nextExpected.Type || results[index].ServerHost != nextExpected.ServerHost {
				gotError = true
				break
			}
		}
	}

	if gotError {
		errMsg := "Mismatched Tasks results.\nExpected:\n\t"
		for _, next := range expected {
			errMsg += fmt.Sprintf("Type: %s ... with ServerHost: %+v\n\t", next.Type, next.ServerHost)
		}
		errMsg += "\n But Got:\n\t"
		for _, next := range results {
			errMsg += fmt.Sprintf("Type: %s ... with ServerHost: %+v\n\t", next.Type, next.ServerHost)
		}

		t.Error(errMsg)
	}
}

func TestGetNode(t *testing.T) {
	FlushTables(t)
	LoadTagFixtures(tagFixtures, t)
	LoadNamedServerFixtures(namedServerFixtures, t)

	dbNode := testNodes["11Kenya"]
	err := PutItem(domain.DataTable, dbNode)
	if err != nil {
		t.Errorf("Error saving Node fixture to db.\n%s", err.Error())
	}

	node, err := GetNode(testNodes["11Kenya"].MacAddr)
	if err != nil {
		t.Errorf("Unexpected error from GetNode. %s", err.Error())
	}

	expected := []domain.Tag{tagFixtures[0], tagFixtures[1]}
	if !areTagsEqual(expected, node.Tags, t) {
		return
	}

	expectedNS := namedServerFixtures[0]
	namedServer := node.Tasks[0].NamedServer // domain.NamedServer{}

	if namedServer.Name != expectedNS.Name {
		t.Errorf(
			"Mismatching Named Server for first Task. Expected Name: %s. But got Name: %s",
			expectedNS.Name,
			namedServer.Name,
		)
	}

}

func TestGetUserByUserID(t *testing.T) {
	FlushTables(t)
	LoadTagFixtures(tagFixtures, t)

	oldTags := []domain.Tag{
		{ID: "tag-000", UID: "000", Name: "Eastern Africa"},    // This name should change
		{ID: "tag-111", UID: "111", Name: "Anglophone Africa"}, // This name should change
		{ID: "tag-999", UID: "999", Name: "Not In DB"},         // This tag should get dropped
	}

	userID := "tommy_tester"
	dbUser := domain.User{
		ID:     "user-" + userID,
		UserID: userID,
		Tags:   oldTags,
	}
	err := PutItem(domain.DataTable, dbUser)
	if err != nil {
		t.Errorf("Error saving User fixture to db.\n%s", err.Error())
		return
	}

	results, err := GetUserByUserID(userID)
	if err != nil {
		t.Errorf("Unexpected error getting user. %s", err.Error())
		return
	}
	if results.UserID != userID {
		t.Errorf("Got wrong user. Expected UserID: %s, \n but got %v", userID, results)
		return
	}
	expected := []domain.Tag{tagFixtures[0], tagFixtures[1]}
	areTagsEqual(expected, results.Tags, t)
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

	// We can't trust the order of the servers
	lowestServer := servers[0]
	if servers[1].ID < lowestServer.ID {
		lowestServer = servers[1]
	}

	results = lowestServer.ID
	expected = 111
	if expected != results {
		t.Errorf("Wrong server ID. Expected: %d. But got: %d", expected, results)
		return
	}

	resultsHost := lowestServer.Host
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

func TestGetTaskLogForRange(t *testing.T) {
	FlushTables(t)

	fixturesInRange := []domain.TaskLogEntry{
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145185,
			Upload:    10.0,
			Download:  20.0,
		},
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145285,
			Upload:    10.0,
			Download:  20.0,
		},
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145385,
			Upload:    10.0,
			Download:  20.0,
		},
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145485,
			Upload:    10.0,
			Download:  20.0,
		},
		{
			ID:        "ping-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145485,
			Upload:    10.0,
			Download:  20.0,
		},
		{
			ID:        "speedTest-11:22:33:44:55:66",
			MacAddr:   "11:22:33:44:55:66",
			Timestamp: 1528145485,
			Upload:    10.0,
			Download:  20.0,
		},
		{
			ID:        "ping-11:22:33:44:55:66",
			MacAddr:   "11:22:33:44:55:66",
			Timestamp: 1528145485,
			Upload:    10.0,
			Download:  20.0,
		},
	}

	fixturesOutOfRange := []domain.TaskLogEntry{
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 2528145085,
			Upload:    10.0,
			Download:  20.0,
		},
	}

	// Load fixtures into db
	for _, fix := range fixturesInRange {
		PutItem(domain.TaskLogTable, fix)
	}
	for _, fix := range fixturesOutOfRange {
		PutItem(domain.TaskLogTable, fix)
	}

	results, err := GetTaskLogForRange(1528145185, 1528145485, "", []string{domain.TaskTypePing, domain.TaskTypeSpeedTest})
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(results) != len(fixturesInRange) {
		t.Error("Not enough results returned, got ", len(results), "expected", len(fixturesInRange))
	}

	// Get just out of range fixtures
	results, err = GetTaskLogForRange(2528145085, 2528145085, "", []string{domain.TaskTypePing, domain.TaskTypeSpeedTest})
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(results) != len(fixturesOutOfRange) {
		t.Error("Not enough results returned, got ", len(results), "expected", len(fixturesOutOfRange))
	}

	// Get just results for MacAddr:   "aa:bb:cc:dd:11:22"
	results, err = GetTaskLogForRange(1528145185, 1528145485, "aa:bb:cc:dd:11:22", []string{domain.TaskTypePing, domain.TaskTypeSpeedTest})
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	for _, entry := range results {
		if entry.MacAddr != "aa:bb:cc:dd:11:22" {
			t.Error("entry returned for wrong mac address. only wanted", "aa:bb:cc:dd:11:22", "got:", entry.MacAddr)
			t.Fail()
		}
	}

	// Get all speedTest results
	results, err = GetTaskLogForRange(1528145185, 1528145485, "", []string{domain.TaskTypeSpeedTest})
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	for _, entry := range results {
		if !strings.HasPrefix(entry.ID, domain.TaskTypeSpeedTest) {
			t.Error("entry returned for wrong type. only wanted", "speedTest", "got:", entry.ID)
			t.Fail()
		}
	}

	// Get all speedTest results for MacAddr: "aa:bb:cc:dd:11:22"
	results, err = GetTaskLogForRange(1528145185, 1528145485, "aa:bb:cc:dd:11:22", []string{domain.TaskTypeSpeedTest})
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	for _, entry := range results {
		if !strings.HasPrefix(entry.ID, domain.TaskTypeSpeedTest) {
			t.Error("entry returned for wrong type. only wanted", "speedTest", "got:", entry.ID)
			t.Fail()
		}
		if entry.MacAddr != "aa:bb:cc:dd:11:22" {
			t.Error("entry returned for wrong mac address. only wanted", "aa:bb:cc:dd:11:22", "got:", entry.MacAddr)
			t.Fail()
		}
	}

}

package speedtestnet

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"net/http"
	"net/http/httptest"
	"testing"
)

// See the host values on these (3333 is missing intentionally)
const ServerListResponse = `<?xml version="1.0" encoding="UTF-8"?>
<settings>
<servers>
 <speedtestnetserver url="http://88.84.191.230/speedtest/upload.php" lat="70.0733" lon="29.7497" name="Vadso" country="Norway" cc="NO" sponsor="Varanger KraftUtvikling AS" id="0000"  url2="http://speedmonster.varangerbynett.no/speedtest/upload.php" host="fine.host.com:8080" />
 <speedtestnetserver url="http://speedtest.nornett.net/speedtest/upload.php" lat="69.9403" lon="23.3106" name="Alta" country="Norway" cc="NO" sponsor="Nornett AS" id="1111"  url2="http://speedtest2.nornett.net/speedtest/upload.php" host="outdated.host.com:8080" />
 <speedtestnetserver url="http://speedo.eltele.no/speedtest/upload.php" lat="69.9403" lon="23.3106" name="Alta" country="Norway" cc="NO" sponsor="Eltele AS" id="2222"  host="good.host.com:8080" />
 <speedtestnetserver url="http://tos.speedtest.as2116.net/speedtest/upload.php" lat="69.6492" lon="18.9553" name="Tromsø" country="Norway" cc="NO" sponsor="Broadnet" id="4444"  host="no.namedserver.com:8080" />
</servers>
</settings>`

func TestGetSTNetServers(t *testing.T) {

	mux := http.NewServeMux()

	testServer := httptest.NewServer(mux)

	respBody := ServerListResponse

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-type", "test/xml")
		w.WriteHeader(200)
		fmt.Fprintf(w, respBody)
	})

	servers, err := GetSTNetServers(testServer.URL)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	expectedLen := 4
	resultsLen := len(servers)
	if resultsLen != 4 {
		t.Errorf("Wrong number of servers. Expected: %d. Got: %d", expectedLen, resultsLen)
		t.Fail()
	}

	expectedIDs := []string{"0000", "1111", "2222", "4444"}
	for index, nextServer := range servers {
		result := nextServer.ServerID
		expected := expectedIDs[index]
		if result != expected {
			t.Errorf("Wrong speedtestnetserver ID at index %d. Expected: %s. Got: %s", index, expected, result)
			t.Fail()
			break
		}
	}
}

func getSTNetServer(serverID string) domain.SpeedTestNetServer {
	server := domain.SpeedTestNetServer{
		ID:          domain.ServerTypeSpeedTestNet + "-" + serverID,
		Host:        "test.url.com:8080",
		Lat:         "11.1",
		Lon:         "22.2",
		Name:        "Sometown",
		Country:     "Anyland",
		CountryCode: "AL",
		ServerID:    serverID,
		Sponsor:     "GTIS",
		URL:         "test.url.com",
		URL2:        "test2.url.com",
	}
	return server
}

func TestServerHasChangedNoChange(t *testing.T) {
	oldServer := getSTNetServer("1111")
	newServer := getSTNetServer("1111")

	results := serverHasChanged(oldServer, newServer)

	if results {
		t.Error("Expected serverHasChanged to say there was no change")
	}
}

func TestServerHasChangedWithChange(t *testing.T) {
	oldServer := getSTNetServer("1111")
	newServer := getSTNetServer("1111")
	newServer.URL2 = "test2b.url.com"

	results := serverHasChanged(oldServer, newServer)

	if !results {
		t.Error("Expected serverHasChanged to say there was a change")
	}
}

type DBClient struct{}

func (d DBClient) DeleteItem(tableAlias, dataType, value string) (bool, error) {
	return true, nil
}

func (d DBClient) PutItem(tableAlias string, item interface{}) error {
	return nil
}

func (d DBClient) ListSpeedTestNetServers() ([]domain.SpeedTestNetServer, error) {
	servers := []domain.SpeedTestNetServer{
		{
			ID:       "speedtestnetserver-st00",
			ServerID: "0000",
			Host:     "fine.host.com:8080",
		},
		{
			ID:       "speedtestnetserver-st11",
			ServerID: "1111",
			Host:     "outdated.host.com:8080",
		},
		{
			ID:       "speedtestnetserver-st22",
			ServerID: "2222",
			Host:     "good.host.com:8080",
		},
		{
			ID:       "speedtestnetserver-st33",
			ServerID: "3333",
			Host:     "missing.server.com:8080",
		},
		{
			ID:       "speedtestnetserver-st44",
			ServerID: "4444",
			Host:     "no.namedserver.com:8080",
		},
	}
	return servers, nil
}

func (d DBClient) ListNamedServers() ([]domain.NamedServer, error) {
	servers := []domain.NamedServer{
		{
			ID:         "namedserver-9999",
			UID:        "9999",
			ServerType: domain.ServerTypeCustom,
		},
		{
			ID:                   "namedserver-ns11",
			UID:                  "ns11",
			ServerType:           domain.ServerTypeSpeedTestNet,
			SpeedTestNetServerID: "1111",
			ServerHost:           "outdated.host.com:8080",
			Name:                 "Outdated Host",
			Description:          "Needs to get its host updated",
			TargetRegion:         "West Africa",
			Notes:                "This named server should have its host value updated.",
		},
		{
			ID:                   "namedserver-ns22",
			UID:                  "ns22",
			ServerType:           domain.ServerTypeSpeedTestNet,
			SpeedTestNetServerID: "2222",
			ServerHost:           "good.host.com:8080",
			Name:                 "Good Host",
			Description:          "No update needed",
			TargetRegion:         "East Africa",
			Notes:                "This named server should not need to be updated.",
		},
		{
			ID:                   "namedserver-ns33",
			UID:                  "ns33",
			ServerType:           domain.ServerTypeSpeedTestNet,
			SpeedTestNetServerID: "3333",
			ServerHost:           "missing.server.com:8080",
			Name:                 "Missing Server",
			Description:          "Has gone stale",
			TargetRegion:         "South Africa",
			Notes:                "This named server has no new matching speedtest.net server",
		},
	}
	return servers, nil
}

func GetTestSTNetServers() ([]domain.SpeedTestNetServer, []domain.SpeedTestNetServer, error) {
	dbCl := DBClient{}

	oldServers, err := dbCl.ListSpeedTestNetServers()
	if err != nil {
		return []domain.SpeedTestNetServer{}, []domain.SpeedTestNetServer{}, err
	}

	// Give one of the new servers a different host
	updatedServer := oldServers[1]
	updatedServer.Host = "updated.host.com"

	newServers := []domain.SpeedTestNetServer{oldServers[0], updatedServer, oldServers[2], oldServers[4]}
	return oldServers, newServers, nil
}

func TestDeleteSTNetServersIfUnused(t *testing.T) {
	dbCl := DBClient{}

	oldServers, newServers, err := GetTestSTNetServers()
	if err != nil {
		t.Errorf("Error getting Server fixture: %s", err.Error())
		return
	}

	namedServers, err := getNamedServersInMap(dbCl)
	if err != nil {
		t.Errorf("Error getting NamedServer fixture: %s", err.Error())
		return
	}

	staleServerIDs, err := deleteSTNetServersIfUnused(oldServers, newServers, namedServers, dbCl)
	if err != nil {
		t.Errorf("Did not expect an error but got: %s", err.Error())
		return
	}

	lenResults := len(staleServerIDs)
	if lenResults != 1 {
		t.Errorf("Expected only one stale Server ID but got: %d", lenResults)
		return
	}

	results := staleServerIDs[0]
	expected := "3333"
	if results != expected {
		t.Errorf("Got wrong stale ServerID. Expected %s, but got %s.", expected, results)
		return
	}

}

func TestUpdateMatchingNamedServer(t *testing.T) {
	dbCl := DBClient{}

	_, newServers, err := GetTestSTNetServers()
	if err != nil {
		t.Errorf("Error getting Server fixture: %s", err.Error())
		return
	}

	namedServers, err := getNamedServersInMap(dbCl)
	if err != nil {
		t.Errorf("Error getting NamedServer fixture: %s", err.Error())
		return
	}

	// only get SpeedTestNet servers
	mappedNamedServers := map[string]domain.NamedServer{}
	for _, namedSrv := range namedServers {
		if namedSrv.ServerType == domain.ServerTypeSpeedTestNet {
			mappedNamedServers[namedSrv.SpeedTestNetServerID] = namedSrv
		}
	}

	// Only one of the new speedtest.net servers has a matching NamedServer that needs to be updated.
	expectedServers := []domain.NamedServer{
		domain.NamedServer{},
		namedServers["1111"],
		domain.NamedServer{},
		domain.NamedServer{},
	}

	for index, targetServer := range newServers[2:3] {
		matchingNamedServer, err := updateMatchingNamedServer(targetServer, mappedNamedServers, dbCl)
		if err != nil {
			t.Errorf("Did not expect an error but got: %s", err.Error())
			return
		}

		expectedServer := expectedServers[index]
		if expectedServer.ID == "" {
			if matchingNamedServer.ID != "" {
				t.Errorf("Did not expect to get a matching NamedServer, but got: %v", matchingNamedServer)
				return
			} else {
				continue
			}
		}

		results := matchingNamedServer.SpeedTestNetServerID
		expected := expectedServer.SpeedTestNetServerID
		if expected != results {
			t.Errorf("Got wrong ServerID. Expected %s, but got %s.", expected, results)
			return
		}
	}

}

func TestUpdateSTNetServers(t *testing.T) {

	dbCl := DBClient{}

	mux := http.NewServeMux()

	testServer := httptest.NewServer(mux)

	// See the host values on these
	respBody := ServerListResponse

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-type", "test/xml")
		w.WriteHeader(200)
		fmt.Fprintf(w, respBody)
	})

	staleServerIDs, err := UpdateSTNetServers(testServer.URL, dbCl)
	if err != nil {
		t.Errorf("Did not expect an error but got: %s", err.Error())
		return
	}

	lenResults := len(staleServerIDs)
	if lenResults != 1 {
		t.Errorf("Expected only one stale Server ID but got: %d", lenResults)
		return
	}

	results := staleServerIDs[0]
	expected := "3333"
	if results != expected {
		t.Errorf("Got wrong stale ServerID. Expected %s, but got %s.", expected, results)
		return
	}

}

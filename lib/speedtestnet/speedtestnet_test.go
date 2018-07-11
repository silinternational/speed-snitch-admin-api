package speedtestnet

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"net/http"
	"net/http/httptest"
	"testing"
)

// See the host values on these ("missing" is missing intentionally)
const ServerListResponse = `<?xml version="1.0" encoding="UTF-8"?>
<settings>
<servers>
<server lat="23.2559" lon="-80.9898" name="Miami" country="United States" cc="US" sponsor="STP" id="fine" host="fine.host.com:8080" />
<server lat="38.6266" lon="-106.6539" name="Denver" country="United States" cc="US" sponsor="Dial" id="updating" host="outdated.host.com:8080" />
<server lat="47.3472" lon="3.3851" name="Paris" country="France" cc="FR" sponsor="Eltele AS" id="good"  host="good.host.com:8080" />
<server lat="46.1428" lon="5.1430" name="Lyon" country="France" cc="FR" sponsor="Broadnet" id="no-named-server"  host="no.namedserver.com:8080" />
</servers>
</settings>`

func loadServerFixtures(fixtures []domain.SpeedTestNetServer) error {
	for _, fix := range fixtures {
		err := db.PutItem(&fix)
		if err != nil {
			return err
		}
	}

	expectedLength := len(fixtures)

	err := db.ListItems(&fixtures, "")
	if err != nil {
		return fmt.Errorf("Error calling list Servers: %s", err.Error())
	}
	if len(fixtures) != expectedLength {
		return fmt.Errorf("Wrong number of server fixtures. Expected: %d. But got: %d", expectedLength, len(fixtures))
	}

	return nil
}

func loadNamedServerFixtures(fixtures []domain.NamedServer) error {
	for _, fix := range fixtures {
		err := db.PutItem(&fix)
		if err != nil {
			return err
		}
	}

	expectedLength := len(fixtures)

	err := db.ListItems(&fixtures, "")
	if err != nil {
		return fmt.Errorf("Error calling list NamedServers: %s", err.Error())
	}
	if len(fixtures) != expectedLength {
		return fmt.Errorf("Wrong number of NamedServer fixtures. Expected: %d. But got: %d.\n%+v", expectedLength, len(fixtures), fixtures)
	}

	return nil
}

func loadCountryFixtures(fixtures []domain.Country) error {
	for _, fix := range fixtures {
		err := db.PutItem(&fix)
		if err != nil {
			return err
		}
	}

	expectedLength := len(fixtures)

	err := db.ListItems(&fixtures, "")
	if err != nil {
		return fmt.Errorf("Error calling list Countries: %s", err.Error())
	}
	if len(fixtures) != expectedLength {
		return fmt.Errorf("Wrong number of Country fixtures. Expected: %d. But got: %d.\n%+v", expectedLength, len(fixtures), fixtures)
	}

	return nil
}

func setUpMuxForServerList(serverListResponse string) *httptest.Server {
	mux := http.NewServeMux()

	testServer := httptest.NewServer(mux)

	if serverListResponse == "" {
		serverListResponse = ServerListResponse
	}
	respBody := serverListResponse

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-type", "test/xml")
		w.WriteHeader(200)
		fmt.Fprintf(w, respBody)
	})

	return testServer
}

func TestGetSTNetServers(t *testing.T) {
	testServer := setUpMuxForServerList("")

	servers, countries, err := GetSTNetServers(testServer.URL)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	if len(servers) != 4 {
		t.Errorf("Wrong number of servers. Expected: 4. Got: %d", len(servers))
	}

	expectedIDs := []string{"fine", "updating", "good", "no-named-server"}
	for _, nextID := range expectedIDs {
		_, ok := servers[nextID]
		if !ok {
			t.Errorf("Results servers are missing a key: %s.\n Got: %v", nextID, servers)
			return
		}
	}

	if len(countries) != 2 {
		t.Errorf("Wrong number of countries. Expected: 2. Got: %d", len(countries))
	}

	expectedCodes := []string{"US", "FR"}
	for _, nextCode := range expectedCodes {
		_, ok := countries[nextCode]
		if !ok {
			t.Errorf("Results countries are missing a key: %s.\n Got: %v", nextCode, servers)
			return
		}
	}
}

func TestDeleteOutdatedSTNetServers(t *testing.T) {
	testutils.ResetDb(t)

	goodServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 1,
		},
		Name:        "New York City, NY",
		Country:     "United States",
		CountryCode: "US",
		Host:        "good.host.ny:8080",
		ServerID:    "1111",
	}

	outdatedServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 2,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "fr",
		Host:        "outdated.host.com:8080",
		ServerID:    "2222",
	}

	// Has a NamedServer, so shouldn't get deleted
	missingServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 3,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "fr",
		Host:        "missing.host.com:8080",
		ServerID:    "3333",
	}

	deleteMeServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 6,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "fr",
		Host:        "deleteme.host.com:8080",
		ServerID:    "6666",
	}

	serverFixtures := []domain.SpeedTestNetServer{
		goodServer,
		outdatedServer,
		missingServer,
		deleteMeServer,
	}

	err := loadServerFixtures(serverFixtures)
	if err != nil {
		t.Errorf("Error loading speedtest.net fixtures.\n%s\n", err.Error())
		return
	}

	namedServerFixtures := []domain.NamedServer{
		{
			Model: gorm.Model{
				ID: 9,
			},
			ServerType: domain.ServerTypeCustom,
			ServerHost: "custom.server.org:8080",
		},
		{
			Model: gorm.Model{
				ID: 1,
			},
			ServerType:           domain.ServerTypeSpeedTestNet,
			SpeedTestNetServerID: 1,
			ServerHost:           goodServer.Host,
			Name:                 "Good Host NY",
			Description:          "No update needed",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			ServerType:         domain.ServerTypeSpeedTestNet,
			SpeedTestNetServer: outdatedServer,
			ServerHost:         "outdated.host.com:8080",
			Name:               "Outdated Host",
			Description:        "Needs to get its host updated",
		},
		{
			Model: gorm.Model{
				ID: 666,
			},
			ServerType:         domain.ServerTypeSpeedTestNet,
			SpeedTestNetServer: missingServer,
			ServerHost:         missingServer.Host,
			Name:               "Missing Server",
			Description:        "Needs to remain as is, since no new server",
		},
	}

	err = loadNamedServerFixtures(namedServerFixtures)
	if err != nil {
		t.Errorf("Error loading NamedServer fixtures.\n%s\n", err.Error())
		return
	}
	namedServers := map[string]domain.NamedServer{}
	for _, nSrv := range namedServerFixtures[1:] {
		namedServers[nSrv.SpeedTestNetServer.ServerID] = nSrv
	}

	var oldServers []domain.SpeedTestNetServer

	err = db.ListItems(&oldServers, "country_code asc")
	if err != nil {
		t.Errorf("Error getting speedtest.net servers from database: %s", err.Error())
	}

	newServers := map[string]domain.SpeedTestNetServer{
		goodServer.ServerID: {
			Name:        goodServer.Name,
			Country:     goodServer.Country,
			CountryCode: goodServer.CountryCode,
			Host:        goodServer.Host,
			ServerID:    goodServer.ServerID,
		},
		outdatedServer.ServerID: {
			Name:        "Modified Server",
			Country:     outdatedServer.Country,
			CountryCode: outdatedServer.CountryCode,
			Host:        "modified.host.com:8080",
			ServerID:    outdatedServer.ServerID,
		},
		"7777": {
			Name:        "Brand New Server",
			Country:     "FR",
			CountryCode: "France",
			Host:        "new.host.com:8080",
			ServerID:    "7777",
		},
	}

	expected := []string{missingServer.ServerID}
	results := deleteOutdatedSTNetServers(oldServers, newServers, namedServers)
	if len(results) != 1 || results[0] != expected[0] {
		t.Errorf("Wrong Stale Server IDs. Expected: %v. \n\tBut Got %v.", expected, results)
	}

	var updatedServers []domain.SpeedTestNetServer
	err = db.ListItems(&updatedServers, "server_id asc")
	if err != nil {
		t.Errorf("Error calling list items to get updated servers: \n%s", err.Error())
		return
	}

	expectedLength := len(serverFixtures) - 1
	if len(updatedServers) != expectedLength {
		t.Errorf("Wrong number of updated servers. Expected: %d. But got: %d", expectedLength, len(updatedServers))
		return
	}

	expectedIDs := []string{goodServer.ServerID, outdatedServer.ServerID, missingServer.ServerID}
	for index, expected := range expectedIDs {
		if updatedServers[index].ServerID != expected {
			updatedIDs := []string{}
			for _, updatedS := range updatedServers {
				updatedIDs = append(updatedIDs, updatedS.ServerID)
			}
			t.Errorf("Bad updated server IDs. \nExpected: %v\n But Got: %+v", expectedIDs, updatedIDs)
			return
		}
	}

}

//func getNamedServerFixtures() []domain.NamedServer {
//	return []domain.NamedServer{
//		{
//			ID:         "namedserver-9999",
//			UID:        "9999",
//			ServerType: domain.ServerTypeCustom,
//			ServerHost: "custom.server.org:8080",
//		},
//		{
//			ID:                   "namedserver-ns11",
//			UID:                  "ns11",
//			ServerType:           domain.ServerTypeSpeedTestNet,
//			SpeedTestNetServerID: "updating",
//			ServerHost:           "outdated.host.com:8080",
//			Name:                 "Outdated Host",
//			Description:          "Needs to get its host updated",
//			Country:              domain.Country{Code: "WA", Name: "West Africa"},
//			Notes:                "This named server should have its host value updated.",
//		},
//		{
//			ID:                   "namedserver-ns22",
//			UID:                  "ns22",
//			ServerType:           domain.ServerTypeSpeedTestNet,
//			SpeedTestNetServerID: "good",
//			ServerHost:           "good.host.com:8080",
//			Name:                 "Good Host",
//			Description:          "No update needed",
//			Country:              domain.Country{Code: "EA", Name: "East Africa"},
//			Notes:                "This named server should not need to be updated.",
//		},
//		{
//			ID:                   "namedserver-ns33",
//			UID:                  "ns33",
//			ServerType:           domain.ServerTypeSpeedTestNet,
//			SpeedTestNetServerID: "missing",
//			ServerHost:           "missing.server.com:8080",
//			Name:                 "Missing Server",
//			Description:          "Has gone stale",
//			Country:              domain.Country{Code: "SA", Name: "South Africa"},
//			Notes:                "This named server has no new matching speedtest.net server",
//		},
//	}
//}
//
//func getSTNetServerListFixtures() []domain.STNetServerList {
//	serverLists := []domain.STNetServerList{
//		{
//			ID:      domain.DataTypeSTNetServerList + "-US",
//			Country: domain.Country{Code: "US", Name: "United States"},
//			Servers: []domain.SpeedTestNetServer{
//				{
//					ServerID:    "fine",
//					CountryCode: "US",
//					Country:     "United States",
//					Host:        "fine.host.com:8080",
//					Name:        "New York",
//				},
//				{
//					ServerID:    "updating",
//					CountryCode: "US",
//					Country:     "United States",
//					Host:        "outdated.host.com:8080",
//					Name:        "Miami",
//				},
//			},
//		},
//		{
//			ID:      domain.DataTypeSTNetServerList + "-FR",
//			Country: domain.Country{Code: "FR", Name: "France"},
//			Servers: []domain.SpeedTestNetServer{
//				{
//					ServerID:    "good",
//					CountryCode: "FR",
//					Country:     "France",
//					Host:        "good.host.com:8080",
//					Name:        "Paris",
//				},
//				{
//					ServerID:    "missing",
//					CountryCode: "FR",
//					Country:     "France",
//					Host:        "missing.server.com:8080",
//					Name:        "Lyon",
//				},
//				{
//					ServerID:    "no-named-server",
//					CountryCode: "FR",
//					Country:     "France",
//					Host:        "no.namedserver.com:8080",
//					Name:        "Paris",
//				},
//			},
//		},
//		{
//			ID:      domain.DataTypeSTNetServerList + "-ZZ",
//			Country: domain.Country{Code: "ZZ", Name: "Annihilated"},
//			Servers: []domain.SpeedTestNetServer{
//				{
//					ServerID:    "zombies",
//					CountryCode: "ZZ",
//					Country:     "Zombieland",
//					Host:        "zombies.got.uscom:8080",
//					Name:        "Gotham",
//				},
//			},
//		},
//	}
//
//	return serverLists
//}

//
//func TestRefreshSTNetServersByCountry(t *testing.T) {
//	deleteSTNetServerLists(t)
//	oldServerLists := getSTNetServerListFixtures()
//	loadSTNetServerLists(oldServerLists, t)
//
//	updatedServer := oldServerLists[0].Servers[1]
//	updatedServer.Host = "updated.host.com"
//
//	newServers := map[string]domain.SpeedTestNetServer{
//		"fine":            oldServerLists[0].Servers[0],
//		"updating":        updatedServer,
//		"good":            oldServerLists[1].Servers[0],
//		"no-named-server": oldServerLists[1].Servers[2],
//	}
//
//	// This changes the entries in the database
//	err := refreshSTNetServers(newServers)
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err.Error())
//		return
//	}
//
//	dbServers, err := db.ListSTNetServerLists()
//	if err != nil {
//		t.Errorf("Unexpected error getting STNetServerLists: %s", err.Error())
//		return
//	}
//
//	expectedLen := 2
//	resultsLen := len(dbServers)
//	if expectedLen != resultsLen {
//		t.Errorf("Bad results. Expected list with %d elements.\n But got %d: %v", expectedLen, resultsLen, dbServers)
//		return
//	}
//
//	// To facilitate checking the results, put them in a map (the servers should be sorted by their Name values)
//	results := map[string][]domain.SpeedTestNetServer{}
//	for _, server := range dbServers {
//		results[server.Country.Code] = server.Servers
//	}
//
//	// Just check the Host values
//	expected := map[string][]string{} // Host
//	expected["US"] = []string{updatedServer.Host, oldServerLists[0].Servers[0].Host}
//	expected["FR"] = []string{oldServerLists[1].Servers[0].Host, oldServerLists[1].Servers[2].Host}
//
//	for id, serverList := range expected {
//		nextResults, ok := results[id]
//		if !ok {
//			t.Errorf("Missing entry with id: %s.\n%v", id, dbServers)
//			return
//		}
//		if len(nextResults) != len(serverList) {
//			t.Errorf("Bad list of servers for id: %s. Expected %v But got %v", id, serverList, nextResults)
//			return
//		}
//		for index := 0; index < len(serverList); index++ {
//			if serverList[index] != nextResults[index].Host {
//				t.Errorf("Bad list of servers for id: %s. Expected %v\n But got %v", id, serverList, nextResults)
//				return
//			}
//		}
//	}
//
//	// Check that the list of countries got saved to the DB
//	countryList, err := db.GetSTNetCountryList()
//	if err != nil {
//		t.Errorf("Unexpected error retrieving country list for speedtest.net servers.\n%s", err.Error())
//		return
//	}
//	resultsList := countryList.Countries
//	expectedList := []domain.Country{{Code: "FR", Name: "France"}, {Code: "US", Name: "United States"}}
//
//	if len(resultsList) != len(expectedList) {
//		t.Errorf("Bad Country List. Expected: %v.\n\t But got: %v", expectedList, resultsList)
//		return
//	}
//
//	if expectedList[0] != resultsList[0] || expectedList[1] != resultsList[1] {
//		t.Errorf("Bad Country List. Expected: %v.\n\t But got: %v", expectedList, resultsList)
//		return
//	}
//}
//
//func TestRefreshSTNetServersByCountryStartEmpty(t *testing.T) {
//	deleteSTNetServerLists(t)
//
//	// We're using the data but not loading them into the db
//	oldServerLists := getSTNetServerListFixtures()
//
//	updatedServer := oldServerLists[0].Servers[1]
//	updatedServer.Host = "updated.host.com"
//
//	newServers := map[string]domain.SpeedTestNetServer{
//		"fine":            oldServerLists[0].Servers[0],
//		"updating":        updatedServer,
//		"good":            oldServerLists[1].Servers[0],
//		"no-named-server": oldServerLists[1].Servers[2],
//	}
//
//	// This changes the entries in the database
//	err := refreshSTNetServers(newServers)
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err.Error())
//		return
//	}
//
//	dbServers, err := db.ListSTNetServerLists()
//	if err != nil {
//		t.Errorf("Unexpected error getting STNetServerLists: %s", err.Error())
//		return
//	}
//
//	expectedLen := 2
//	resultsLen := len(dbServers)
//	if expectedLen != resultsLen {
//		t.Errorf("Bad results. Expected list with %d elements.\n But got %d: %v", expectedLen, resultsLen, dbServers)
//		return
//	}
//
//	// To facilitate checking the results, put them in a map and sort the servers by their Host values
//	results := map[string][]domain.SpeedTestNetServer{}
//	for _, server := range dbServers {
//		countryServers := server.Servers
//
//		sort.Slice(countryServers, func(i, j int) bool {
//			return countryServers[i].Host < countryServers[j].Host
//		})
//		results[server.Country.Code] = countryServers
//	}
//
//	// Just check the Host values
//	expected := map[string][]string{} // Host
//	expected["US"] = []string{oldServerLists[0].Servers[0].Host, updatedServer.Host}
//	expected["FR"] = []string{oldServerLists[1].Servers[0].Host, oldServerLists[1].Servers[2].Host}
//
//	for id, serverList := range expected {
//		nextResults, ok := results[id]
//		if !ok {
//			t.Errorf("Missing entry with id: %s.\n%v", id, dbServers)
//			return
//		}
//		if len(nextResults) != len(serverList) {
//			t.Errorf("Bad list of servers for id: %s. Expected %v But got %v", id, serverList, nextResults)
//			return
//		}
//		for index := 0; index < len(serverList); index++ {
//			if serverList[index] != nextResults[index].Host {
//				t.Errorf("Bad list of servers for id: %s. Expected %v\n But got %v", id, serverList, nextResults)
//				return
//			}
//		}
//	}
//}

func TestGetSTNetNamedServers(t *testing.T) {
	testutils.ResetDb(t)

	goodServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 1,
		},
		Name:        "New York City, NY",
		Country:     "United States",
		CountryCode: "US",
		Host:        "good.host.ny:8080",
		ServerID:    "1111",
	}

	outdatedServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 2,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "fr",
		Host:        "outdated.host.com:8080",
		ServerID:    "2222",
	}

	serverFixtures := []domain.SpeedTestNetServer{
		goodServer,
		outdatedServer,
	}

	err := loadServerFixtures(serverFixtures)
	if err != nil {
		t.Errorf("Error loading speedtest.net fixtures.\n%s\n", err.Error())
		return
	}

	namedServerFixtures := []domain.NamedServer{
		{
			Model: gorm.Model{
				ID: 9,
			},
			ServerType: domain.ServerTypeCustom,
			ServerHost: "custom.server.org:8080",
		},
		{
			Model: gorm.Model{
				ID: 1,
			},
			ServerType:           domain.ServerTypeSpeedTestNet,
			SpeedTestNetServerID: 1,
			Name:                 "Good Host NY",
			Description:          "No update needed",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			ServerType:           domain.ServerTypeSpeedTestNet,
			SpeedTestNetServerID: 2,
			Name:                 "Outdated Host",
			Description:          "Needs to get its host updated",
		},
	}

	err = loadNamedServerFixtures(namedServerFixtures)
	if err != nil {
		t.Errorf("Error loading NamedServer fixtures.\n%s\n", err.Error())
		return
	}
	expectedServers := map[string]domain.NamedServer{}
	for _, nSrv := range namedServerFixtures[1:] {
		expectedServers[nSrv.SpeedTestNetServer.ServerID] = nSrv
	}

	allResults, err := getSTNetNamedServers()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	results := len(allResults)
	expected := 2
	if results != expected {
		t.Errorf("Bad results. Expected list with %d elements.\n But got %d: %v", expected, results, allResults)
		return
	}

	expectedIDs := []string{goodServer.ServerID, outdatedServer.ServerID}

	for _, expected := range expectedIDs {
		_, ok := allResults[expected]
		if !ok {
			resultsIDs := []string{}
			for serverID := range allResults {
				resultsIDs = append(resultsIDs, serverID)
			}
			t.Errorf("Bad NamedServer serverIDs. \nExpected: %v\n But Got: %+v", expectedIDs, resultsIDs)
			return
		}
	}

}

//func TestGetSTNetServersToKeep(t *testing.T) {
//	deleteNamedServers(t)
//	deleteSTNetServerLists(t)
//
//	namedServers := loadAndGetNamedServers(t)
//
//	oldServerLists := getSTNetServerListFixtures()
//	oldServerLists = oldServerLists[0:2] // don't keep the last country (ZZ)
//
//	// Give one of the new servers a different host
//	updatedServer := oldServerLists[0].Servers[1]
//	updatedServer.Host = "brand.new.host.com"
//
//	newServers := map[string]domain.SpeedTestNetServer{
//		"fine":            oldServerLists[0].Servers[0],
//		"updating":        updatedServer,
//		"good":            oldServerLists[1].Servers[0],
//		"no-named-server": oldServerLists[1].Servers[2],
//	}
//
//	serversToKeep, staleServerIDs := getSTNetServersToKeep(oldServerLists, newServers, namedServers)
//
//	// First check serversToKeep
//	expectedServers := map[string]domain.SpeedTestNetServer{
//		"fine":            oldServerLists[0].Servers[0],
//		"updating":        updatedServer,
//		"good":            oldServerLists[1].Servers[0],
//		"missing":         oldServerLists[1].Servers[1],
//		"no-named-server": oldServerLists[1].Servers[2],
//	}
//
//	expectedLen := len(expectedServers)
//	resultsLen := len(serversToKeep)
//
//	if expectedLen != resultsLen {
//		t.Errorf("Bad serversToKeep length. Expected %d, but got %d.\n %v", expectedLen, resultsLen, serversToKeep)
//		return
//	}
//
//	for id, expected := range expectedServers {
//		oneResult, ok := serversToKeep[id]
//		if !ok {
//			t.Errorf("Bad serversToKeep. Missing entry for ID: %s.\n Got %v", id, serversToKeep)
//			return
//		}
//
//		if expected != oneResult {
//			t.Errorf("Bad serversToKeep. \n  Expected: %v\n  But got: %v", expectedServers, serversToKeep)
//			return
//		}
//	}
//
//	// Now check staleServerIDs
//	expectedIDs := []string{"missing"}
//
//	if len(staleServerIDs) != 1 {
//		t.Errorf("Bad staleServerIDs. Expected %v, but got %v.", expectedIDs, staleServerIDs)
//		return
//	}
//
//	if expectedIDs[0] != staleServerIDs[0] {
//		t.Errorf("Bad staleServerIDs. Expected %v, but got %v.", expectedIDs, staleServerIDs)
//	}
//}
//
//func TestUpdateNamedServers(t *testing.T) {
//	deleteNamedServers(t)
//	namedServers := loadAndGetNamedServers(t)
//
//	// One Host is different (for "updating") and one is no longer there ("missing")
//	serversToKeep := map[string]domain.SpeedTestNetServer{
//		"fine":            {ServerID: "fine", Host: "fine.host.com:8080"},
//		"updating":        {ServerID: "updating", Host: "brand.new.host.com:8080"},
//		"good":            {ServerID: "good", Host: "good.host.com:8080"},
//		"no-named-server": {ServerID: "no-named-server", Host: "no.namedserver.com:8080"},
//	}
//
//	err := updateNamedServers(serversToKeep, namedServers)
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err.Error())
//		return
//	}
//
//	results, err := getSTNetNamedServers()
//	if err != nil {
//		t.Errorf("Unexpected error getting NamedServers out of the db: %s", err.Error())
//		return
//	}
//
//	expected := map[string]string{
//		"updating": serversToKeep["updating"].Host, // This is a change
//		"good":     serversToKeep["good"].Host,
//		"missing":  "missing.server.com:8080", // updateNamedServers does not delete any
//	}
//
//	if len(results) != len(expected) {
//		t.Errorf("Bad Results. Expected %d NamedServers from db,\n but got %d: %v", len(expected), len(results), results)
//		return
//	}
//
//	for id, expectedHost := range expected {
//		oneResult, ok := results[id]
//		if !ok {
//			t.Errorf("Bad Results. Missing entry for ID: %s.\n Got %v", id, results)
//			return
//		}
//
//		if expectedHost != oneResult.ServerHost {
//			t.Errorf("Bad Host for NamedServer: %s. Expected: %s, but got: %s", id, expectedHost, oneResult.ServerHost)
//		}
//	}
//}
//
//// This test doesn't check the results in depth, since the other tests do that.
//func TestUpdateSTNetServers(t *testing.T) {
//	deleteNamedServers(t)
//	namedServers := getNamedServerFixtures()
//	loadNamedServers(namedServers, t)
//
//	deleteSTNetServerLists(t)
//	loadSTNetServerLists(getSTNetServerListFixtures(), t)
//
//	testServer := setUpMuxForServerList()
//
//	staleServerIDs, err := UpdateSTNetServers(testServer.URL)
//	if err != nil {
//		t.Errorf("Did not expect an error but got: %s", err.Error())
//		return
//	}
//
//	// First, check the staleServerIDs
//	lenResults := len(staleServerIDs)
//	if lenResults != 1 {
//		t.Errorf("Expected one stale Server ID but got: %d", lenResults)
//		return
//	}
//
//	results := staleServerIDs[0]
//	expected := "missing"
//	if results != expected {
//		t.Errorf("Got wrong stale ServerID. Expected %s, but got %s.", expected, results)
//		return
//	}
//
//	// Second, check the number of NamedServers in the database
//	dbNamedServers, err := db.ListNamedServers()
//	if err != nil {
//		t.Errorf("Unexpected error getting NamedServers from db: %s", err.Error())
//		return
//	}
//
//	if len(namedServers) != len(dbNamedServers) {
//		t.Errorf("Bad Named Servers left in db. \n\tExpected: %v\n\t But got: %v", namedServers, dbNamedServers)
//		return
//	}
//
//	// Third, check the number of STNetServerLists in the database
//	stNetServerLists, err := db.ListSTNetServerLists()
//	if err != nil {
//		t.Errorf("Unexpected error getting STNetServerLists from db: %s", err.Error())
//		return
//	}
//
//	expectedLen := 2
//	resultsLen := len(stNetServerLists)
//
//	if expectedLen != resultsLen {
//		t.Errorf(
//			"Wrong number of STNetServerLists. Expected: %d, but got: %d.\n\t%v",
//			expectedLen,
//			resultsLen,
//			stNetServerLists,
//		)
//	}
//}

func TestUpdateCountries(t *testing.T) {
	testutils.ResetDb(t)

	goodCountry := domain.Country{
		Model: gorm.Model{
			ID: 1,
		},
		Code: "GC",
		Name: "Good Country",
	}

	deleteMeCountry := domain.Country{
		Model: gorm.Model{
			ID: 2,
		},
		Code: "DC",
		Name: "Delete Me Country",
	}

	outdatedCountry := domain.Country{
		Model: gorm.Model{
			ID: 3,
		},
		Code: "OC",
		Name: "Outdated Country",
	}

	fixtures := []domain.Country{goodCountry, deleteMeCountry, outdatedCountry}
	err := loadCountryFixtures(fixtures)
	if err != nil {
		t.Errorf("Error loading Country fixtures.\n%s\n", err.Error())
		return
	}

	newCountries := map[string]domain.Country{
		"GC": goodCountry,
		"OC": domain.Country{Code: outdatedCountry.Code, Name: "Updated Country"},
	}

	updateCountries(newCountries)

	var updatedCountries []domain.Country
	err = db.ListItems(&updatedCountries, "code asc")
	if err != nil {
		t.Errorf("Error calling list Countries: %s", err.Error())
		return
	}

	if len(updatedCountries) != 2 {
		t.Errorf("Wrong number of Country entries after update. Expected: 2. But Got: %d.\n%+v",
			len(updatedCountries),
			updatedCountries,
		)
		return
	}

	expectedCountries := []domain.Country{goodCountry, newCountries["OC"]}

	for index := range []int{0, 1} {
		if expectedCountries[index].Code != updatedCountries[index].Code ||
			expectedCountries[index].Name != updatedCountries[index].Name {
			t.Errorf("Updated Country results not as expected at index %d. \nExpected: %+v.\n But Got: %+v.",
				index,
				expectedCountries[index],
				updatedCountries[index],
			)
			return
		}
	}
}

func TestUpdateSTNetServers(t *testing.T) {
	testutils.ResetDb(t)
	serverListResponse := `<?xml version="1.0" encoding="UTF-8"?>
<settings>
<servers>
<server lat="23.2" lon="-80.9" name="Miami" country="United States" cc="US" sponsor="STP" id="7777" host="new.host.com:8080" />
<server lat="38.6" lon="-106.6" name="New York City, NY" country="United States" cc="US" sponsor="Dial" id="1111" host="good.host.com:8080" />
</servers>
</settings>`

	testServer := setUpMuxForServerList(serverListResponse)

	goodServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 1,
		},
		Name:        "New York City, NY",
		Country:     "United States",
		CountryCode: "US",
		Host:        "good.host.ny:8080",
		ServerID:    "1111",
	}

	// Has a NamedServer, so shouldn't get deleted
	missingServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 2,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "fr",
		Host:        "missing.host.com:8080",
		ServerID:    "2222",
	}

	deleteMeServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 6,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "fr",
		Host:        "deleteme.host.com:8080",
		ServerID:    "6666",
	}

	serverFixtures := []domain.SpeedTestNetServer{
		goodServer,
		missingServer,
		deleteMeServer,
	}

	err := loadServerFixtures(serverFixtures)
	if err != nil {
		t.Errorf("Error loading speedtest.net fixtures.\n%s\n", err.Error())
		return
	}

	namedServerFixtures := []domain.NamedServer{
		{
			Model: gorm.Model{
				ID: 9,
			},
			ServerType: domain.ServerTypeCustom,
			ServerHost: "custom.server.org:8080",
		},
		{
			Model: gorm.Model{
				ID: 1,
			},
			ServerType:         domain.ServerTypeSpeedTestNet,
			SpeedTestNetServer: goodServer,
			ServerHost:         goodServer.Host,
			Name:               "Good Host NY",
			Description:        "No update needed",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			ServerType:         domain.ServerTypeSpeedTestNet,
			SpeedTestNetServer: missingServer,
			ServerHost:         missingServer.Host,
			Name:               "Missing Server",
			Description:        "Needs to remain as is, since no new server",
		},
	}

	err = loadNamedServerFixtures(namedServerFixtures)
	if err != nil {
		t.Errorf("Error loading NamedServer fixtures.\n%s\n", err.Error())
		return
	}

	usCountry := domain.Country{
		Model: gorm.Model{
			ID: 1,
		},
		Code: "US",
		Name: "United States",
	}

	deleteMeCountry := domain.Country{
		Model: gorm.Model{
			ID: 2,
		},
		Code: "DC",
		Name: "Delete Me Country",
	}

	fixtures := []domain.Country{usCountry, deleteMeCountry}
	err = loadCountryFixtures(fixtures)
	if err != nil {
		t.Errorf("Error loading Country fixtures.\n%s\n", err.Error())
		return
	}

	staleServerIDs, err := UpdateSTNetServers(testServer.URL)
	if err != nil {
		t.Errorf("Unexpected error ... %s", err.Error())
		return
	}

	expected := []string{missingServer.ServerID}
	if len(staleServerIDs) != 1 || staleServerIDs[0] != expected[0] {
		t.Errorf("Bad staleServerID results. Expected: %v.\n\t But got: %v", expected, staleServerIDs)
	}

	var updatedServers []domain.SpeedTestNetServer
	err = db.ListItems(&updatedServers, "server_id asc")
	if err != nil {
		t.Errorf("Error calling list items to get updated servers: \n%s", err.Error())
		return
	}

	expectedLength := len(serverFixtures)
	if len(updatedServers) != expectedLength {
		t.Errorf("Wrong number of updated servers. Expected: %d. But got: %d\n%+v", expectedLength, len(updatedServers), updatedServers)
		return
	}

	expectedIDs := []string{goodServer.ServerID, missingServer.ServerID, "7777"}
	for index, expected := range expectedIDs {
		if updatedServers[index].ServerID != expected {
			updatedIDs := []string{}
			for _, updatedS := range updatedServers {
				updatedIDs = append(updatedIDs, updatedS.ServerID)
			}
			t.Errorf("Bad updated server IDs. \nExpected: %v\n But Got: %+v", expectedIDs, updatedIDs)
			return
		}
	}
}

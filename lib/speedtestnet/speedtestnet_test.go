package speedtestnet

//
//import (
//	"fmt"
//	"github.com/silinternational/speed-snitch-admin-api"
//	"github.com/silinternational/speed-snitch-admin-api/db"
//	"net/http"
//	"net/http/httptest"
//	"sort"
//	"testing"
//)
//
//// See the host values on these ("missing" is missing intentionally)
//const ServerListResponse = `<?xml version="1.0" encoding="UTF-8"?>
//<settings>
//<servers>
// <server lat="23.2559" lon="-80.9898" name="Miami" country="United States" cc="US" sponsor="STP" id="fine" host="fine.host.com:8080" />
// <server lat="38.6266" lon="-106.6539" name="Denver" country="United States" cc="US" sponsor="Dial" id="updating" host="outdated.host.com:8080" />
// <server lat="47.3472" lon="3.3851" name="Paris" country="France" cc="FR" sponsor="Eltele AS" id="good"  host="good.host.com:8080" />
// <server lat="46.1428" lon="5.1430" name="Lyon" country="France" cc="FR" sponsor="Broadnet" id="no-named-server"  host="no.namedserver.com:8080" />
//</servers>
//</settings>`
//
//func deleteNamedServers(t *testing.T) {
//	items, err := db.ListNamedServers()
//	if err != nil {
//		t.Errorf("Could not get list of NamedServers for test preparations. Got error: %s", err.Error())
//		t.Fail()
//	}
//
//	for _, nextItem := range items {
//		success, err := db.DeleteItem(domain.DataTable, domain.DataTypeNamedServer, nextItem.UID)
//		if err != nil || !success {
//			t.Errorf("Could not delete NamedServer from db for test preparations.  UID: %s. %s", nextItem.UID, err.Error())
//			t.Fail()
//		}
//	}
//
//	items, err = db.ListNamedServers()
//	if err != nil {
//		t.Errorf("Could not get list of NamedServers for test preparations. Got error: %s", err.Error())
//		t.Fail()
//	}
//
//	if len(items) > 0 {
//		t.Errorf("Did not succeed in deleting NamedServers for test preparations. Got %d items", len(items))
//		t.Fail()
//	}
//}
//
//func deleteSTNetServerLists(t *testing.T) {
//	items, err := db.ListSTNetServerLists()
//	if err != nil {
//		t.Errorf("Could not get list of SpeedTestNetServers for test preparations. Got error: %s", err.Error())
//		t.Fail()
//	}
//
//	for _, nextItem := range items {
//		success, err := db.DeleteItem(domain.DataTable, domain.DataTypeSTNetServerList, nextItem.Country.Code)
//		if err != nil || !success {
//			t.Errorf("Could not delete SpeedTestNetServers from db for test preparations.  Country Code: %s. %s", nextItem.Country.Code, err.Error())
//			t.Fail()
//		}
//	}
//
//	items, err = db.ListSTNetServerLists()
//	if err != nil {
//		t.Errorf("Could not get list of SpeedTestNetServers for test preparations. Got error: %s", err.Error())
//		t.Fail()
//	}
//
//	if len(items) > 0 {
//		t.Errorf("Did not succeed in deleting SpeedTestNetServers for test preparations. Got %d items", len(items))
//		t.Fail()
//	}
//}
//
//func loadNamedServers(namedServers []domain.NamedServer, t *testing.T) {
//	for _, namedServer := range namedServers {
//		err := db.PutItem(domain.DataTable, &namedServer)
//		if err != nil {
//			t.Errorf(
//				"Could not load NamedServer into db for test preparations. UID: %s. \nError: %s",
//				namedServer.UID,
//				err.Error(),
//			)
//			t.Fail()
//		}
//	}
//}
//
//func loadAndGetNamedServers(t *testing.T) map[string]domain.NamedServer {
//	loadNamedServers(getNamedServerFixtures(), t)
//	namedServers, err := getSTNetNamedServers()
//	if err != nil {
//		t.Errorf("Could not get NamedServers from the db to prepare for the test.\n%s", err.Error())
//		t.Fail()
//	}
//
//	return namedServers
//}
//
//func loadSTNetServerLists(stNetServerLists []domain.STNetServerList, t *testing.T) {
//	for _, server := range stNetServerLists {
//		err := db.PutItem(domain.DataTable, &server)
//		if err != nil {
//			t.Errorf(
//				"Could not load NamedServer into db for test preparations. ID: %s. \nError: %s",
//				server.ID,
//				err.Error(),
//			)
//			t.Fail()
//		}
//	}
//}
//
//func setUpMuxForServerList() *httptest.Server {
//	mux := http.NewServeMux()
//
//	testServer := httptest.NewServer(mux)
//
//	respBody := ServerListResponse
//
//	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
//		w.Header().Set("Content-type", "test/xml")
//		w.WriteHeader(200)
//		fmt.Fprintf(w, respBody)
//	})
//
//	return testServer
//}
//
//func TestGetSTNetServers(t *testing.T) {
//	testServer := setUpMuxForServerList()
//
//	servers, err := GetSTNetServers(testServer.URL)
//	if err != nil {
//		t.Errorf(err.Error())
//		t.Fail()
//	}
//
//	expectedLen := 4
//	resultsLen := len(servers)
//	if resultsLen != 4 {
//		t.Errorf("Wrong number of servers. Expected: %d. Got: %d", expectedLen, resultsLen)
//		t.Fail()
//	}
//
//	expectedIDs := []string{"fine", "updating", "good", "no-named-server"}
//	for _, nextID := range expectedIDs {
//		_, ok := servers[nextID]
//		if !ok {
//			t.Errorf("Results servers are missing a key: %s.\n Got: %v", nextID, servers)
//			return
//		}
//	}
//}
//
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
//func TestHasAHostChangedFalse(t *testing.T) {
//	oldList := []domain.SpeedTestNetServer{
//		{ServerID: "0000", Host: "acme.com:8080"},
//		{ServerID: "1111", Host: "beta.com:8080"},
//	}
//
//	newList := []domain.SpeedTestNetServer{
//		{ServerID: "0000", Host: "acme.com:8080"},
//		{ServerID: "1111", Host: "beta.com:8080"},
//	}
//
//	if hasAHostChanged(oldList, newList) {
//		t.Errorf("Didn't expect to see that a host had changed, but did.")
//	}
//}
//
//func TestHasAHostChangedTrue(t *testing.T) {
//	oldList := []domain.SpeedTestNetServer{
//		{ServerID: "0000", Host: "acme.com:8080"},
//		{ServerID: "1111", Host: "beta.com:8080"},
//	}
//
//	newList := []domain.SpeedTestNetServer{
//		{ServerID: "0000", Host: "acme-new.com:8080"},
//		{ServerID: "1111", Host: "beta.com:8080"},
//	}
//
//	if !hasAHostChanged(oldList, newList) {
//		t.Errorf("Expected to see that a host had changed, but didn't.")
//	}
//}
//
//func TestHasAHostChangedTrueDifferentLengths(t *testing.T) {
//	oldList := []domain.SpeedTestNetServer{
//		{ServerID: "1111", Host: "beta.com:8080"},
//	}
//
//	newList := []domain.SpeedTestNetServer{
//		{ServerID: "0000", Host: "acme.com:8080"},
//		{ServerID: "1111", Host: "beta.com:8080"},
//	}
//
//	if !hasAHostChanged(oldList, newList) {
//		t.Errorf("Expected to see that a host had changed, but didn't.")
//	}
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
//	err := refreshSTNetServersByCountry(newServers)
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
//	err := refreshSTNetServersByCountry(newServers)
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
//
//func TestGetSTNetNamedServers(t *testing.T) {
//	deleteNamedServers(t)
//	loadNamedServers(getNamedServerFixtures(), t)
//
//	allResults, err := getSTNetNamedServers()
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err.Error())
//		return
//	}
//
//	results := len(allResults)
//	expected := 3
//	if results != expected {
//		t.Errorf("Bad results. Expected list with %d elements.\n But got %d: %v", expected, results, allResults)
//		return
//	}
//
//	expectedServerIDs := []string{"updating", "good", "missing"}
//
//	for _, expected := range expectedServerIDs {
//		_, ok := allResults[expected]
//		if !ok {
//			t.Errorf("Missing NamedServer entry with UID: %s. \n Got: %v", expected, allResults)
//			return
//		}
//	}
//
//}
//
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

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
		"GC": domain.Country{Code: goodCountry.Code, Name: goodCountry.Name},
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

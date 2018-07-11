package speedtestnet

import (
	"encoding/xml"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"io/ioutil"
	"net/http"
	"os"
)

// GetSTNetServers requests the list of SpeedTestNet servers via http and returns them in a map of structs
//  with the ServerID's as keys
func GetSTNetServers(serverURL string) (map[string]domain.SpeedTestNetServer, map[string]domain.Country, error) {
	var outerXML domain.STNetServerSettings

	servers := map[string]domain.SpeedTestNetServer{}
	countries := map[string]domain.Country{} // by country code

	resp, err := http.Get(serverURL)
	if err != nil {
		return servers, countries, fmt.Errorf("Error making http Get for SpeedTestNet servers: \n\t%s", err.Error())
	}

	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return servers, countries, fmt.Errorf("Error reading SpeedTestNet servers from http response: \n\t%s", err.Error())
	}

	xml.Unmarshal(respBytes, &outerXML)

	for _, nextServerList := range outerXML.ServerLists {
		for _, nextServer := range nextServerList.Servers {
			servers[nextServer.ServerID] = nextServer
			country := domain.Country{
				Code: nextServer.CountryCode,
				Name: nextServer.Country,
			}
			countries[country.Code] = country
		}
	}

	return servers, countries, nil
}

func deleteOutdatedSTNetServers(
	oldServers []domain.SpeedTestNetServer,
	newServers map[string]domain.SpeedTestNetServer,
	namedServers map[string]domain.NamedServer,
) []string {

	staleServerIDs := []string{}

	for _, oldie := range oldServers {
		serverID := oldie.ServerID
		_, exists := newServers[serverID] // If there is a matching new server, nothing to do
		if exists {
			continue
		}

		_, exists = namedServers[serverID] // If there is a matching NamedServer, remember it and don't delete the old server
		if exists {
			staleServerIDs = append(staleServerIDs, serverID)
			continue
		}

		// Since there is no matching new server or NamedServer, delete the old server
		oldie, err := db.GetSpeedTestNetServerByServerID(serverID)
		if err != nil {
			domain.ErrorLogger.Println("\nError finding outdated speedtest.net server to delete. ServerID: ", serverID)
		} else {
			err = db.DeleteItem(&oldie, oldie.ID)
			if err != nil {
				domain.ErrorLogger.Println("\nError deleting outdated speedtest.net server. ServerID: ", serverID, "\n", err)
			}
		}
	}
	return staleServerIDs
}

// getSTNetNamedServers returns a map with the NamedServers in the database that
// have a ServerType of speedtestnet.  The keys are the SpeedTestNet ServerID's.
func getSTNetNamedServers() (map[string]domain.NamedServer, error) {
	mappedNamedServers := map[string]domain.NamedServer{}

	var namedServers []domain.NamedServer
	err := db.ListItems(&namedServers, "name asc")
	if err != nil {
		return mappedNamedServers, fmt.Errorf("Error getting speedtest.net servers from database: %s", err.Error())
	}

	for _, namedSrv := range namedServers {
		if namedSrv.ServerType == domain.ServerTypeSpeedTestNet {
			mappedNamedServers[namedSrv.SpeedTestNetServer.ServerID] = namedSrv
		}
	}

	return mappedNamedServers, nil
}

// updateNamedServers updates the NamedServer entries with a new Host value
//  based on data from the new set of SpeedTestNetServers
func updateNamedServers(
	newServers map[string]domain.SpeedTestNetServer,
	namedServers map[string]domain.NamedServer,
) {

	for serverID, namedServer := range namedServers {
		if namedServer.ServerType != domain.ServerTypeSpeedTestNet {
			continue
		}

		newServer, exists := newServers[serverID]
		if !exists {
			continue
		}

		// Found a match, so check if it needs to be modified
		if namedServer.ServerHost != newServer.Host {
			namedServer.ServerHost = newServer.Host
			err := db.PutItem(&namedServer)
			if err != nil {
				errMsg := fmt.Sprintf(
					"\nError updating Named Server %s with new host: %s",
					namedServer.ID,
					err.Error(),
				)

				domain.ErrorLogger.Println(errMsg)
			}
		}
	}
}

func updateCountries(newCountries map[string]domain.Country) {

	var oldCountries []domain.Country
	err := db.ListItems(&oldCountries, "code asc")
	if err != nil {
		domain.ErrorLogger.Println("\nError getting Countries from database: ", err.Error())
		return
	}

	for countryCode, country := range newCountries {
		err := db.PutItem(&country)
		if err != nil {
			domain.ErrorLogger.Println("\nError updating Country with code ", countryCode)
		}
	}

	// Delete old countries in the db that don't match a new country
	for _, oldCountry := range oldCountries {
		_, exists := newCountries[oldCountry.Code]
		if exists {
			continue
		}

		err = db.DeleteItem(&oldCountry, oldCountry.ID)
		if err != nil {
			domain.ErrorLogger.Println("\nError deleting Country with ID %v.", oldCountry.ID)
		}
	}
}

// UpdateSTNetServers returns a list of the IDs of speedtest.net servers that are no longer available
//   but have a matching Named Server.  Also,
//     -- it updates (in the database) all SpeedTestNetServer entries with matching new ones.
//     -- it updates NamedServers entries that match a new SpeedTestNetServer which has a new Host value.
//     -- it leaves in the db SpeedTestNetServer entries that do not match a new one but are associated with
//         a NamedServer entry
func UpdateSTNetServers(serverURL string) ([]string, error) {
	var oldSTNetServers []domain.SpeedTestNetServer
	err := db.ListItems(&oldSTNetServers, "country_code asc")
	if err != nil {
		return []string{}, fmt.Errorf("Error getting speedtest.net servers from database: %s", err.Error())
	}

	fmt.Fprintf(os.Stdout, "\nFound %v old servers", len(oldSTNetServers))

	newServers, newCountries, err := GetSTNetServers(serverURL)
	if err != nil {
		return []string{}, fmt.Errorf("Error getting new speedtest.net servers: %s", err.Error())
	}
	fmt.Fprintf(os.Stdout, "\nFound %v new servers", len(newServers))

	namedServers, err := getSTNetNamedServers()
	if err != nil {
		return []string{}, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
	}
	fmt.Fprintf(os.Stdout, "\nFound %v named servers", len(namedServers))

	// Delete old SpeedTestNetServers that don't have a matching new one and get a list of the NamedServers that don't have a match anymore
	staleServerIDs := deleteOutdatedSTNetServers(oldSTNetServers, newServers, namedServers)
	fmt.Fprintf(os.Stdout, "\nFound %v outdated servers that still have a matching NamedServer", len(staleServerIDs))

	// Where necessary, make the Named Servers' Host values match those in the corresponding new SpeedTestNetServers
	updateNamedServers(newServers, namedServers)

	updateCountries(newCountries)

	// Save changes to the new set of SpeedTestNetServers
	for serverID, server := range newServers {
		err := db.PutItem(&server)
		if err != nil && err != gorm.ErrRecordNotFound {
			domain.ErrorLogger.Println(
				"\nCould not save or update speedtest.net server in db. ServerID: ",
				serverID,
				"\n",
				err.Error(),
			)
		}
	}

	return staleServerIDs, nil
}

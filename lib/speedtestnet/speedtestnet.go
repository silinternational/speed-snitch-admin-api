package speedtestnet

//
//import (
//	"encoding/xml"
//	"fmt"
//	"github.com/silinternational/speed-snitch-admin-api"
//	"github.com/silinternational/speed-snitch-admin-api/db"
//	"io/ioutil"
//	"net/http"
//	"os"
//	"sort"
//)
//
//// GetSTNetServers requests the list of SpeedTestNet servers via http and returns them in a map of structs
////  with the ServerID's as keys
//func GetSTNetServers(serverURL string) (map[string]domain.SpeedTestNetServer, error) {
//	var settings domain.STNetServerSettings
//
//	servers := map[string]domain.SpeedTestNetServer{}
//
//	resp, err := http.Get(serverURL)
//	if err != nil {
//		return servers, fmt.Errorf("Error making http Get for SpeedTestNet servers: \n\t%s", err.Error())
//	}
//
//	defer resp.Body.Close()
//
//	respBytes, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return servers, fmt.Errorf("Error reading SpeedTestNet servers from http response: \n\t%s", err.Error())
//	}
//
//	xml.Unmarshal(respBytes, &settings)
//
//	for _, nextServerList := range settings.ServerLists {
//		for _, nextServer := range nextServerList.Servers {
//			servers[nextServer.ServerID] = nextServer
//		}
//	}
//
//	return servers, nil
//}
//
//// getSTNetServersToKeep ...
////   - returns a map of Servers that are either still valid (available as a speedtest.net server) or are
////     not valid but are associated with a NamedServer.
////   - returns a slice of the ServerIDs that are no longer valid but that
////     are associated with a NamedServer
//func getSTNetServersToKeep(
//	oldServerLists []domain.STNetServerList,
//	newServers map[string]domain.SpeedTestNetServer,
//	namedServers map[string]domain.NamedServer,
//) (map[string]domain.SpeedTestNetServer, []string) {
//
//	staleServerIDs := []string{}
//	serversToKeep := newServers
//
//	for _, oldList := range oldServerLists {
//		for _, oldServer := range oldList.Servers {
//			_, oldServerHasANewCounterpart := newServers[oldServer.ServerID]
//			_, oldServerHasANamedServer := namedServers[oldServer.ServerID]
//
//			// if the old server has a new counterpart, it's already in the output map
//			// if the old server is not associated with a named server, do NOT add it back in
//			if oldServerHasANewCounterpart || !oldServerHasANamedServer {
//				continue
//			}
//
//			// At this point, the old server does not have a new counterpart but it is
//			//  associated with a NamedServer
//			staleServerIDs = append(staleServerIDs, oldServer.ServerID)
//			serversToKeep[oldServer.ServerID] = oldServer
//		}
//	}
//
//	return serversToKeep, staleServerIDs
//}
//
//// hasAHostChanged checks if
////   -- the length of the two lists is different or
////   -- if any of the old serverIDs are missing from the new servers or
////   -- if any of the matching Host values are different
//func hasAHostChanged(oldServers, newServers []domain.SpeedTestNetServer) bool {
//	if len(oldServers) != len(newServers) {
//		return true
//	}
//
//	newHosts := map[string]string{}
//	for _, server := range newServers {
//		newHosts[server.ServerID] = server.Host
//	}
//
//	oldHosts := map[string]string{}
//	for _, server := range oldServers {
//		oldHosts[server.ServerID] = server.Host
//	}
//
//	for oldID, oldHost := range oldHosts {
//		newHost, ok := newHosts[oldID]
//		if !ok || oldHost != newHost {
//			return true
//		}
//	}
//
//	return false
//}
//
//// refreshSTNetServersByCountry takes the new SpeedTestNetServers and groups them by country.
////  If any of the old country groupings are not represented in the new ones, it deletes them.
////  It updates all the other country groupings of servers with the new data.
////  *** It also saves a list of the countries as one row in the database
//func refreshSTNetServersByCountry(servers map[string]domain.SpeedTestNetServer) error {
//	countries := map[string]domain.Country{}
//
//	// This groups the new servers by country (based on country code)
//	// and builds a map of just the countries themselves
//	groupedServers := map[string]domain.STNetServerList{}
//	for _, server := range servers {
//		// Keep track of the countries separately
//		country := domain.Country{Code: server.CountryCode, Name: server.Country}
//		countries[country.Code] = country
//		_, ok := groupedServers[server.CountryCode]
//		if !ok {
//			groupedServers[server.CountryCode] = domain.STNetServerList{
//				Country: country,
//				Servers: []domain.SpeedTestNetServer{server},
//			}
//		} else {
//			// It appears that Go requires this extra processing to avoid compile errors
//			updatedEntry := groupedServers[server.CountryCode]
//			updatedEntry.Servers = append(updatedEntry.Servers, server)
//			groupedServers[server.CountryCode] = updatedEntry
//		}
//	}
//
//	// Delete or update the old server lists first
//	oldServers, err := db.ListSTNetServerLists()
//	if err != nil {
//		return fmt.Errorf("Error trying to get the SpeedTestNetServerLists from the db: %s", err.Error())
//	}
//
//	oldCountries := map[string]bool{}
//	countriesAdded := 0
//	countriesUpdated := 0
//
//	for _, oldServerList := range oldServers {
//		oldCountries[oldServerList.Country.Code] = true
//		newServerList, ok := groupedServers[oldServerList.Country.Code]
//		// If the country is still represented in the new data, and if it has a significant change, update it
//		if ok {
//			if hasAHostChanged(oldServerList.Servers, newServerList.Servers) {
//				sortedServers := newServerList.Servers
//				// Sort ascending by country name
//				sort.Slice(sortedServers, func(i, j int) bool {
//					return sortedServers[i].Name < sortedServers[j].Name
//				})
//				newServerList.Servers = sortedServers
//
//				newServerList.ID = domain.DataTypeSTNetServerList + "-" + oldServerList.Country.Code
//				err := db.PutItem(domain.DataTable, &newServerList)
//				if err != nil {
//					return fmt.Errorf("Error trying to update SpeedTestNetServerList, %s, in the db: %s", newServerList.ID, err.Error())
//				}
//				countriesUpdated++
//			}
//			// If the country is no longer represented in the new data, delete it
//		} else {
//			_, err := db.DeleteItem(domain.DataTable, domain.DataTypeSTNetServerList, oldServerList.Country.Code)
//			if err != nil {
//				return fmt.Errorf("Error trying to delete SpeedTestNetServerList, %s, from the db: %s", oldServerList.ID, err.Error())
//			}
//			fmt.Fprintf(os.Stdout, "Deleting SpeedTestNetServerList entry for country code %s\n", oldServerList.Country.Code)
//		}
//	}
//
//	fmt.Fprintf(os.Stdout, "Updated server lists for %d countries.\n", countriesUpdated)
//
//	// Now if there are new server lists that were not in the db add them.
//	for countryCode, newServerList := range groupedServers {
//		_, ok := oldCountries[countryCode]
//		if !ok {
//			newServerList.ID = domain.DataTypeSTNetServerList + "-" + countryCode
//			err := db.PutItem(domain.DataTable, &newServerList)
//			if err != nil {
//				return fmt.Errorf("Error trying to add SpeedTestNetServerList for %s, in the db: %s", countryCode, err.Error())
//			}
//			countriesAdded++
//		}
//	}
//
//	fmt.Fprintf(os.Stdout, "Added server lists for %d countries.\n", countriesAdded)
//
//	countryList := []domain.Country{}
//	for _, country := range countries {
//		countryList = append(countryList, country)
//	}
//
//	// Sort ascending by country name
//	sort.Slice(countryList, func(i, j int) bool {
//		return countryList[i].Name < countryList[j].Name
//	})
//
//	stNetCountryList := domain.STNetCountryList{
//		ID:        domain.DataTypeSTNetCountryList + "-" + domain.STNetCountryListUID,
//		Countries: countryList,
//	}
//
//	err = db.PutItem(domain.DataTable, &stNetCountryList)
//	if err != nil {
//		return fmt.Errorf("Error trying to update the list of countries for speedtest.net servers.\n%s", err.Error())
//	}
//	return nil
//}
//
//// getSTNetNamedServers returns a map with the NamedServers in the database that
//// have a ServerType of speedtestnet.  The keys are the SpeedTestNet ServerID's.
//func getSTNetNamedServers() (map[string]domain.NamedServer, error) {
//	mappedNamedServers := map[string]domain.NamedServer{}
//	namedServers, err := db.ListNamedServers()
//	if err != nil {
//		return map[string]domain.NamedServer{}, err
//	}
//
//	for _, namedSrv := range namedServers {
//		if namedSrv.ServerType == domain.ServerTypeSpeedTestNet {
//			mappedNamedServers[namedSrv.SpeedTestNetServerID] = namedSrv
//		}
//	}
//
//	return mappedNamedServers, nil
//}
//
//// updateNamedServers updates the NamedServer entries with a new Host value
////  based on data from the new set of SpeedTestNetServers
//func updateNamedServers(
//	serversToKeep map[string]domain.SpeedTestNetServer,
//	namedServers map[string]domain.NamedServer,
//) error {
//
//	for id, namedServer := range namedServers {
//		newServer, ok := serversToKeep[id]
//
//		if !ok {
//			continue
//		}
//
//		// Found a match, so check if it needs to be modified
//		if namedServer.ServerHost != newServer.Host {
//			namedServer.ServerHost = newServer.Host
//			err := db.PutItem(domain.DataTable, &namedServer)
//			if err != nil {
//				return fmt.Errorf(
//					"Error updating Named Server %s with new host: %s",
//					namedServer.ID,
//					err.Error(),
//				)
//			}
//		}
//	}
//	return nil
//}
//
//// UpdateSTNetServers returns a list of the IDs of speedtest.net servers that are no longer available
////   but have a matching Named Server.  Also,
////     -- it replaces (in the database) all SpeedTestNetServer entries with the new ones but keeps old
////        ones that still are referenced by a NamedServer
////     -- it updates NamedServers entries that match a new SpeedTestNetServer but have with a new Host value.
//func UpdateSTNetServers(serverURL string) ([]string, error) {
//	oldServerLists, err := db.ListSTNetServerLists()
//	if err != nil {
//		return []string{}, fmt.Errorf("Error getting speedtest.net servers from database: %s", err.Error())
//	}
//
//	serverCount := 0
//	for _, serverList := range oldServerLists {
//		serverCount += len(serverList.Servers)
//	}
//
//	fmt.Fprintf(os.Stdout, "Found %v old server lists containing %v servers\n", len(oldServerLists), serverCount)
//
//	newServers, err := GetSTNetServers(serverURL)
//	if err != nil {
//		return []string{}, err
//	}
//	fmt.Fprintf(os.Stdout, "Found %v new servers\n", len(newServers))
//
//	namedServers, err := getSTNetNamedServers()
//	if err != nil {
//		return []string{}, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
//	}
//	fmt.Fprintf(os.Stdout, "Found %v named servers\n", len(namedServers))
//
//	// Get an updated set of SpeedTestNetServers and a list of the NamedServers that don't have a match anymore
//	serversToKeep, staleServerIDs := getSTNetServersToKeep(oldServerLists, newServers, namedServers)
//	fmt.Fprintf(os.Stdout, "Found %v stale servers\n", len(staleServerIDs))
//
//	// Where necessary, make the Named Servers' Host values match those in the corresponding new SpeedTestNetServers
//	err = updateNamedServers(serversToKeep, namedServers)
//	if err != nil {
//		return staleServerIDs, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
//	}
//
//	// Save the new set of SpeedTestNetServers
//	err = refreshSTNetServersByCountry(serversToKeep)
//
//	return staleServerIDs, nil
//}

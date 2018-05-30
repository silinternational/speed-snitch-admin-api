package speedtestnet

import (
	"encoding/xml"
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"io/ioutil"
	"net/http"
	"os"
)

// see https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_BatchWriteItem.html
func getBatchRequest(requests []string) string {
	requestStart := fmt.Sprintf(`{
  "RequestItems": {
	"%s": [`,
		domain.DataTypeSpeedTestNetServer,
	)

	wholeRequest := requestStart
	lastIndex := len(requests) - 1

	for index, request := range requests {
		wholeRequest += request
		if index < lastIndex {
			wholeRequest += ","
		}
		wholeRequest += `
`
	}

	wholeRequest += `
    ]
  }
}`
	return wholeRequest
}

// GetSTNetServers requests the list of SpeedTestNet servers via http and returns them in a map of structs
//  with the ServerID's as keys
func GetSTNetServers(serverURL string) (map[string]domain.SpeedTestNetServer, error) {
	var settings domain.STNetServerSettings

	servers := map[string]domain.SpeedTestNetServer{}

	resp, err := http.Get(serverURL)
	if err != nil {
		return servers, fmt.Errorf("Error making http Get for SpeedTestNet servers: \n\t%s", err.Error())
	}

	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return servers, fmt.Errorf("Error reading SpeedTestNet servers from http response: \n\t%s", err.Error())
	}

	xml.Unmarshal(respBytes, &settings)

	for _, nextServerList := range settings.ServerLists {
		for _, nextServer := range nextServerList.Servers {
			servers[nextServer.ServerID] = nextServer
		}
	}

	return servers, nil
}

// getSTNetServersToKeep ...
//   - returns a map of Servers that are either still valid or are
//     not valid but are associated with a NamedServer.
//   - returns a slice of the ServerIDs that are no longer valid but that
//     are associated with a NamedServer
func getSTNetServersToKeep(
	oldServerLists []domain.STNetServerList,
	newServers map[string]domain.SpeedTestNetServer,
	namedServers map[string]domain.NamedServer,
) (map[string]domain.SpeedTestNetServer, []string) {

	staleServerIDs := []string{}
	serversToKeep := newServers

	for _, oldList := range oldServerLists {
		for _, oldServer := range oldList.Servers {
			_, oldServerHasANewCounterpart := newServers[oldServer.ServerID]
			_, oldServerHasANamedServer := namedServers[oldServer.ServerID]

			// if the old server has a new counterpart, it's already in the output map
			// if the old server is not associated with a named server, do NOT add it back in
			if oldServerHasANewCounterpart || !oldServerHasANamedServer {
				continue
			}

			// At this point, the old server does not have a new counterpart but it is
			//  associated with a NamedServer
			staleServerIDs = append(staleServerIDs, oldServer.ServerID)
			serversToKeep[oldServer.ServerID] = oldServer
		}
	}

	return serversToKeep, staleServerIDs
}

// refreshSTNetServersByCountry takes the new SpeedTestNetServers and groups them by country.
//  If any of the old country groupings are not represented in the new ones, it deletes them.
//  It updates all the other country groupings of servers with the new data.
func refreshSTNetServersByCountry(servers map[string]domain.SpeedTestNetServer) error {

	groupedServers := map[string]domain.STNetServerList{}
	for _, server := range servers {
		_, ok := groupedServers[server.CountryCode]
		if !ok {
			groupedServers[server.CountryCode] = domain.STNetServerList{
				Country: domain.Country{Code: server.CountryCode, Name: server.Country},
				Servers: []domain.SpeedTestNetServer{server},
			}
		} else {
			// It appears that Go requires this extra processing to avoid compile errors
			updatedEntry := groupedServers[server.CountryCode]
			updatedEntry.Servers = append(updatedEntry.Servers, server)
			groupedServers[server.CountryCode] = updatedEntry
		}
	}

	oldServers, err := db.ListSpeedTestNetServers()
	if err != nil {
		return fmt.Errorf("Error trying to get the SpeedTestNetServerLists from the db: %s", err.Error())
	}

	for _, serverList := range oldServers {
		newServerList, ok := groupedServers[serverList.Country.Code]
		// If the country is still represented in the new data, update it
		if ok {
			newServerList.ID = domain.DataTypeSpeedTestNetServerList + "-" + serverList.Country.Code
			err := db.PutItem(domain.DataTable, &newServerList)
			if err != nil {
				return fmt.Errorf("Error trying to update SpeedTestNetServerList, %s, in the db: %s", newServerList.ID, err.Error())
			}
			// If the country is no longer represented in the new data, delete it
		} else {
			_, err := db.DeleteItem(domain.DataTable, domain.DataTypeSpeedTestNetServerList, serverList.Country.Code)
			if err != nil {
				return fmt.Errorf("Error trying to delete SpeedTestNetServerList, %s, from the db: %s", serverList.ID, err.Error())
			}
			fmt.Fprintf(os.Stdout, "Deleting SpeedTestNetServerList entry for country code %s\n", serverList.Country.Code)
		}
	}

	return nil
}

// getSTNetNamedServers returns a map with the NamedServers in the database that
// have a ServerType of speedtestnet.  The keys are the SpeedTestNet ServerID's.
func getSTNetNamedServers() (map[string]domain.NamedServer, error) {
	mappedNamedServers := map[string]domain.NamedServer{}
	namedServers, err := db.ListNamedServers()
	if err != nil {
		return map[string]domain.NamedServer{}, err
	}

	for _, namedSrv := range namedServers {
		if namedSrv.ServerType == domain.ServerTypeSpeedTestNet {
			mappedNamedServers[namedSrv.SpeedTestNetServerID] = namedSrv
		}
	}

	return mappedNamedServers, nil
}

// updateNamedServers updates the NamedServer entries with a new Host value
//  based on data from the new set of SpeedTestNetServers
func updateNamedServers(
	serversToKeep map[string]domain.SpeedTestNetServer,
	namedServers map[string]domain.NamedServer,
) error {

	for id, namedServer := range namedServers {
		newServer, ok := serversToKeep[id]

		if !ok {
			continue
		}

		// Found a match, so check if it needs to be modified
		if namedServer.ServerHost != newServer.Host {
			namedServer.ServerHost = newServer.Host
			err := db.PutItem(domain.DataTable, &namedServer)
			if err != nil {
				return fmt.Errorf(
					"Error updating Named Server %s with new host: %s",
					namedServer.ID,
					err.Error(),
				)
			}
		}
	}
	return nil
}

// UpdateSTNetServers returns a list of the IDs of speedtest.net servers that are no longer available
//   but have a matching Named Server.  Also,
//     -- it replaces (in the database) all SpeedTestNetServer entries with the new ones but keeps old
//        ones that still are referenced by a NamedServer
//     -- it updates NamedServers entries that match a new SpeedTestNetServer with a different HOST value.
func UpdateSTNetServers(serverURL string) ([]string, error) {

	// Figure out which speedtest.net servers from the database are no longer valid
	oldServerLists, err := db.ListSpeedTestNetServers()
	if err != nil {
		return []string{}, fmt.Errorf("Error getting speedtest.net servers from database: %s", err.Error())
	}

	oldServers := map[string]domain.SpeedTestNetServer{}
	for _, serverList := range oldServerLists {
		for _, server := range serverList.Servers {
			oldServers[server.ServerID] = server
		}
	}

	fmt.Fprintf(os.Stdout, "Found %v old servers\n", len(oldServers))

	newServers, err := GetSTNetServers(serverURL)
	if err != nil {
		return []string{}, err
	}
	fmt.Fprintf(os.Stdout, "Found %v new servers\n", len(newServers))

	namedServers, err := getSTNetNamedServers()
	if err != nil {
		return []string{}, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
	}
	fmt.Fprintf(os.Stdout, "Found %v named servers\n", len(namedServers))

	// Get an updated set of SpeedTestNetServers
	serversToKeep, staleServerIDs := getSTNetServersToKeep(oldServerLists, newServers, namedServers)
	fmt.Fprintf(os.Stdout, "Found %v stale servers\n", len(staleServerIDs))

	// Where necessary, make the Named Servers' Host values match those in the corresponding new SpeedTestNetServers
	err = updateNamedServers(serversToKeep, namedServers)
	if err != nil {
		return staleServerIDs, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
	}

	// Save the new set of SpeedTestNetServers
	err = refreshSTNetServersByCountry(serversToKeep)

	return staleServerIDs, nil
}

package speedtestnet

import (
	"encoding/xml"
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"io/ioutil"
	"net/http"
)

// GetSTNetServers requests the list of SpeedTestNet servers via http and returns them in a list of structs
//  Normally use the domain.SpeedTestNetServerURL as the serverURL
func GetSTNetServers(serverURL string) ([]domain.SpeedTestNetServer, error) {
	var settings domain.STNetServerSettings

	var servers []domain.SpeedTestNetServer

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
			servers = append(servers, nextServer)
		}
	}

	return servers, nil
}

type dbClient interface {
	DeleteItem(string, string, string) (bool, error)
	PutItem(string, interface{}) error
	ListSpeedTestNetServers() ([]domain.SpeedTestNetServer, error)
	ListNamedServers() ([]domain.NamedServer, error)
}

// deleteSTNetServersIfUnused returns a list of the ServerIDs that are referenced by a NamedServer but which
//   no longer appear in the new list of speedtest.net servers.  Also, it deletes (from the database) any old
//   SpeedTestNetServers which do not match an entry in the list of new speedtest.net servers nor are
//   referenced by a NamedServer
func deleteSTNetServersIfUnused(
	oldServers, newServers []domain.SpeedTestNetServer,
	namedServers []domain.NamedServer,
	db dbClient,
) ([]string, error) {

	staleServerIDs := []string{}

	for _, oldServer := range oldServers {
		foundMatch := false
		for _, newServer := range newServers {
			if oldServer.ServerID == newServer.ServerID {
				foundMatch = true
				break
			}
		}

		// If the old server matches a new server, ignore it - it hasn't been deleted
		if foundMatch {
			continue
		}

		inUse := false

		// Found a deleted server, now check if it's in use
		for _, namedServer := range namedServers {
			// If there is a namedServer that refers to a deleted server, keep track of it.
			if namedServer.ServerType == domain.ServerTypeSpeedTestNet && namedServer.SpeedTestNetServerID == oldServer.ServerID {
				staleServerIDs = append(staleServerIDs, oldServer.ServerID)
				inUse = true
				break
			}
		}

		// If there is no new server or namedServer that matches the old server, delete the old server from the db
		if !inUse {
			_, err := db.DeleteItem(domain.DataTable, domain.DataTypeSpeedTestNetServer, oldServer.ServerID)

			if err != nil {
				return []string{}, fmt.Errorf("Error deleting old SpeedTestNetServer %s: %s", oldServer.ID, err.Error())
			}
		}
	}

	return staleServerIDs, nil
}

func serverHasChanged(oldServer, newServer domain.SpeedTestNetServer) bool {

	return (oldServer.Lat != newServer.Lat ||
		oldServer.Lon != newServer.Lon ||
		oldServer.Name != newServer.Name ||
		oldServer.Host != newServer.Host ||
		oldServer.Country != newServer.Country ||
		oldServer.CountryCode != newServer.CountryCode ||
		oldServer.Sponsor != newServer.Sponsor ||
		oldServer.URL != newServer.URL ||
		oldServer.URL2 != newServer.URL2)
}

func updateMatchingNamedServer(
	newServer domain.SpeedTestNetServer,
	namedServers []domain.NamedServer,
	db dbClient,
) (domain.NamedServer, error) {

	var updatedServer domain.NamedServer
	for _, namedServer := range namedServers {
		// Skip those that don't match
		if namedServer.ServerType != domain.ServerTypeSpeedTestNet ||
			namedServer.SpeedTestNetServerID != newServer.ServerID {
			continue
		}

		// Found a match, so check if it needs to be modified
		if namedServer.ServerHost != newServer.Host {
			namedServer.ServerHost = newServer.Host
			err := db.PutItem(domain.DataTable, &namedServer)
			if err != nil {
				return domain.NamedServer{}, fmt.Errorf("Error updating Named Server %s with new host: %s", namedServer.ID, err.Error())
			}
			updatedServer = namedServer
		}
		break
	}
	return updatedServer, nil
}

// UpdateSTNetServers returns a list of the IDs of speedtest.net servers that are no longer available
//   and have a matching Named Server.  Also,
//     -- it deletes (from the database) any old SpeedTestNetServer entries which do not match an entry in
//        the list of new speedtest.net servers nor are referenced by a NamedServer
//     -- it updates old SpeedTestNetServer entries that do match an entry in the new list and updates the
//        matching NamedServers.
func UpdateSTNetServers(serverURL string, db dbClient) ([]string, error) {

	// Figure out which speedtest.net servers have been deleted
	oldServers, err := db.ListSpeedTestNetServers()
	if err != nil {
		return []string{}, fmt.Errorf("Error getting speedtest.net servers from database: %s", err.Error())
	}

	newServers, err := GetSTNetServers(serverURL)
	if err != nil {
		return []string{}, err
	}

	namedServers, err := db.ListNamedServers()
	if err != nil {
		return []string{}, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
	}

	staleServerIDs, err := deleteSTNetServersIfUnused(oldServers, newServers, namedServers, db)

	for _, newServer := range newServers {

		// Find a match in the database and update it as well as a matching NamedServer
		for _, oldServer := range oldServers {
			if newServer.ServerID != oldServer.ServerID {
				continue
			}
			// If there are differences between old and new, update matching NamedServer
			if serverHasChanged(oldServer, newServer) {
				_, err := updateMatchingNamedServer(newServer, namedServers, db)
				if err != nil {
					return []string{}, err
				}
			}
		}
	}

	return staleServerIDs, nil
}

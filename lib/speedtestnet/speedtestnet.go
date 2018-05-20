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

type Client struct{}

func (c Client) DeleteItem(tableAlias, dataType, value string) (bool, error) {
	return db.DeleteItem(tableAlias, dataType, value)
}

func (c Client) PutItem(tableAlias string, item interface{}) error {
	return db.PutItem(tableAlias, item)
}

func (c Client) ListSpeedTestNetServers() ([]domain.SpeedTestNetServer, error) {
	return db.ListSpeedTestNetServers()
}

func (c Client) ListNamedServers() ([]domain.NamedServer, error) {
	return db.ListNamedServers()
}

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
	namedServers map[string]domain.NamedServer,
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

		_, ok := namedServers[oldServer.ServerID]
		if ok {
			staleServerIDs = append(staleServerIDs, oldServer.ServerID)
		} else {
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

func getNamedServersInMap(db dbClient) (map[string]domain.NamedServer, error) {
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

func updateMatchingNamedServer(
	newServer domain.SpeedTestNetServer,
	namedServers map[string]domain.NamedServer,
	db dbClient,
) (domain.NamedServer, error) {

	var updatedServer domain.NamedServer

	namedServer, ok := namedServers[newServer.ServerID]

	// If no match found, nothing to do
	if !ok {
		return updatedServer, nil
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
	fmt.Fprintf(os.Stdout, "Found %v old servers", len(oldServers))

	newServers, err := GetSTNetServers(serverURL)
	if err != nil {
		return []string{}, err
	}
	fmt.Fprintf(os.Stdout, "Found %v new servers", len(newServers))

	namedServers, err := getNamedServersInMap(db)
	if err != nil {
		return []string{}, fmt.Errorf("Error getting Named Servers from database: %s", err.Error())
	}
	fmt.Fprintf(os.Stdout, "Found %v named servers", len(namedServers))

	staleServerIDs, err := deleteSTNetServersIfUnused(oldServers, newServers, namedServers, db)
	fmt.Fprintf(os.Stdout, "Found %v stale servers", len(staleServerIDs))

	// Make a map of the Old Servers for quicker access and avoiding extra checks in a nested loop
	mappedOldServers := map[string]domain.SpeedTestNetServer{}
	for _, oldServer := range oldServers {
		mappedOldServers[oldServer.ServerID] = oldServer
	}

	for _, newServer := range newServers {
		oldServer, ok := mappedOldServers[newServer.ServerID]
		if !ok {
			newServer.ID = domain.DataTypeSpeedTestNetServer + "-" + newServer.ServerID
			err = db.PutItem(domain.DataTable, newServer)
			if err != nil {
				return []string{}, fmt.Errorf("problem with server: %v, error: %s", newServer, err.Error())
			}
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

	return staleServerIDs, nil
}

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

type ServerKey struct {
	ID        string
	Timestamp int64
}

func getDeleteRequest(serverKey ServerKey) string {
	return fmt.Sprintf(`
  {
    "DeleteRequest": {
      "Key": {
        "ID": {"S": "%s"},
        "Timestamp": {"N": "%d"}
      }
    }
  }`,
		serverKey.ID,
		serverKey.Timestamp,
	)
}

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

type Client struct{}

func (c Client) DeleteItem(tableAlias, dataType, value string) (bool, error) {
	return db.DeleteItem(tableAlias, dataType, value)
}

func (c Client) PutItem(tableAlias string, item interface{}) error {
	return db.PutItem(tableAlias, item)
}

func (c Client) ListSpeedTestNetServers() ([]domain.STNetServerList, error) {
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
	ListSpeedTestNetServers() ([]domain.STNetServerList, error)
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
	serverIDsToDelete := []ServerKey{}

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
			serverIDsToDelete = append(
				serverIDsToDelete,
				ServerKey{ID: oldServer.ServerID, Timestamp: oldServer.Timestamp},
			)
			//_, err := db.DeleteItem(domain.DataTable, domain.DataTypeSpeedTestNetServer, oldServer.ServerID)
			//
			//if err != nil {
			//	return []string{}, fmt.Errorf("Error deleting old SpeedTestNetServer %s: %s", oldServer.ID, err.Error())
			//}
		}
	}

	for startIndex := 0; ; startIndex += 25 {
		index := startIndex
		if index >= len(serverIDsToDelete) {
			break
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
		oldServer.Sponsor != newServer.Sponsor)
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

	fmt.Fprintf(os.Stdout, "Found %v old servers", len(oldServerLists))

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

	staleServerIDs, err := deleteSTNetServersIfUnused(oldServerLists, newServers, namedServers, db)
	fmt.Fprintf(os.Stdout, "Found %v stale servers", len(staleServerIDs))

	// Make a map of the Old Servers for quicker access and avoiding extra checks in a nested loop
	mappedOldServers := map[string]domain.SpeedTestNetServer{}
	for _, oldServer := range oldServerLists {
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

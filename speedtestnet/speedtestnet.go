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

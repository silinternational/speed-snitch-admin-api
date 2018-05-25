package main

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"testing"
)

func TestGetCountriesFromSTNetServerLists(t *testing.T) {

	lists := []domain.STNetServerList{
		{
			ID:      "usa",
			Country: domain.Country{Code: "US", Name: "United States"},
			Servers: []domain.SpeedTestNetServer{
				{
					Name:        "New York City, NY",
					Country:     "United States",
					CountryCode: "US",
					Host:        "nyc.speedtest.sbcglobal.net:8080",
					ServerID:    "10390",
				},
				{
					Name:        "Miami, FL",
					Country:     "United States",
					CountryCode: "US",
					Host:        "stosat-pomp-01.sys.comcast.net:8080",
					ServerID:    "1779",
				},
			},
		},
		{
			ID:      "fra",
			Country: domain.Country{Code: "FR", Name: "France"},
			Servers: []domain.SpeedTestNetServer{
				{
					Name:        "Paris",
					Country:     "France",
					CountryCode: "FR",
					Host:        "paris1.speedtest.orange.fr:8080",
					ServerID:    "5559",
				},
				{
					Name:        "Massy",
					Country:     "France",
					CountryCode: "FR",
					Host:        "massy.testdebit.info:8080",
					ServerID:    "2231",
				},
			},
		},
	}

	results, err := getCountriesFromSTNetServerLists(lists)
	if err != nil {
		t.Errorf("Got unexpected error: %s", err.Error())
		return
	}

	expected := []domain.Country{
		{Name: "France", Code: "FR"},
		{Name: "United States", Code: "US"},
	}

	if len(results) != len(expected) {
		t.Errorf("Bad results. Expected %v\n But got %v", expected, results)
		return
	}

	if results[0] != expected[0] || results[1] != expected[1] {
		t.Errorf("Bad results. Expected %v\n But got %v", expected, results)
		return
	}

}

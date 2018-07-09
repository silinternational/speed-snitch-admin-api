package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"strings"
	"testing"
)

func loadServerFixtures(fixtures []domain.SpeedTestNetServer) error {
	for _, fix := range fixtures {
		err := db.PutItem(&fix)
		if err != nil {
			return fmt.Errorf("Error loading fixtures: \n%s", err.Error())
		}
	}

	expectedLength := len(fixtures)

	err := db.ListItems(&fixtures, "")
	if err != nil {
		return fmt.Errorf("Error calling list items: %s", err.Error())
	}
	if len(fixtures) != expectedLength {
		return fmt.Errorf("Wrong number of server fixtures. Expected: %d. But got: %d", expectedLength, len(fixtures))
	}

	return nil
}

func TestViewServer(t *testing.T) {
	testutils.ResetDb(t)

	serverInFocus := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 2,
		},
		Name:        "Paris",
		Country:     "France",
		CountryCode: "FR",
		Host:        "paris1.speedtest.orange.fr:8080",
		ServerID:    "5559",
	}

	serverFixtures := []domain.SpeedTestNetServer{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name:        "New York City, NY",
			Country:     "United States",
			CountryCode: "US",
			Host:        "nyc.speedtest.sbcglobal.net:8080",
			ServerID:    "10390",
		},
		serverInFocus,
		{
			Model: gorm.Model{
				ID: 3,
			},
			Name:        "Miami, FL",
			Country:     "United States",
			CountryCode: "US",
			Host:        "stosat-pomp-01.sys.comcast.net:8080",
			ServerID:    "1779",
		},
		{
			Model: gorm.Model{
				ID: 4,
			},
			Name:        "Massy",
			Country:     "France",
			CountryCode: "FR",
			Host:        "massy.testdebit.info:8080",
			ServerID:    "2231",
		},
	}

	err := loadServerFixtures(serverFixtures)
	if err != nil {
		t.Error(err)
		return
	}

	userFixtures := []domain.User{
		testutils.SuperAdmin,
		testutils.AdminUser,
	}

	for _, fix := range userFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	method := "GET"

	// Test that using an invalid version id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/speedtestnetserver/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	response, err := viewServer(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned viewing version, expected 404, got", response.StatusCode, response.Body)
		return
	}

	// View server
	req = events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/speedtestnetserver/2",
		PathParameters: map[string]string{
			"id": "2",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err = viewServer(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results := response.Body
	if !strings.Contains(results, serverInFocus.ServerID) {
		t.Errorf("viewServer did not include the serverInFocus. Got:\n%s\n", results)
	}

}

func TestListCountries(t *testing.T) {
	testutils.ResetDb(t)

	countryFixtures := []domain.Country{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "United States",
			Code: "US",
		},
		{

			Model: gorm.Model{
				ID: 2,
			},
			Name: "France",
			Code: "FR",
		},
	}

	for _, fix := range countryFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Errorf("Error loading fixtures: \n%s", err.Error())
			return
		}
	}

	expectedLength := len(countryFixtures)

	err := db.ListItems(&countryFixtures, "")
	if err != nil {
		t.Errorf("Error calling list items: %s", err.Error())
		return
	}
	if len(countryFixtures) != expectedLength {
		t.Errorf("Wrong number of server fixtures. Expected: %d. But got: %d", expectedLength, len(countryFixtures))
		return
	}

	userFixtures := []domain.User{
		testutils.SuperAdmin,
		testutils.AdminUser,
	}

	for _, fix := range userFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	method := "GET"

	// Test that using an invalid version id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod:     method,
		Path:           "/speedtestnetserver/country",
		PathParameters: map[string]string{},
		Headers:        testutils.GetSuperAdminReqHeader(),
	}

	response, err := listCountries(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned viewing version, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results := response.Body
	if !strings.Contains(results, "France") || !strings.Contains(results, "United States") {
		t.Errorf("listCountries did not include the fixtures. Got:\n%s\n", results)
	}
}

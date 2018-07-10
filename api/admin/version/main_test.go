package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"strings"
	"testing"
)

func getVersionsCheckLength(expectedLength int) ([]domain.Version, error) {
	versions := []domain.Version{}
	err := db.ListItems(&versions, "number asc")
	if err != nil {
		return versions, fmt.Errorf("Error calling list versions: %s", err.Error())
	}
	if len(versions) != expectedLength {
		return versions, fmt.Errorf("Wrong number of versions. Expected: %d. But got: %d", expectedLength, len(versions))
	}
	return versions, nil
}

func TestDeleteVersion(t *testing.T) {
	testutils.ResetDb(t)

	deleteMeVersion := domain.Version{
		Model: gorm.Model{
			ID: 1,
		},
		Number:      "6.6.6",
		Description: "This version is to be deleted",
	}

	keepMeVersion := domain.Version{
		Model: gorm.Model{
			ID: 2,
		},
		Number:      "3.3.3",
		Description: "This tag is NOT to be deleted",
	}

	userFixtures := []domain.User{
		testutils.SuperAdmin,
		testutils.AdminUser,
	}

	versionFixtures := []domain.Version{
		deleteMeVersion,
		keepMeVersion,
	}

	for _, fix := range userFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, fix := range versionFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	versions, err := getVersionsCheckLength(2)
	if err != nil {
		t.Error(err)
		return
	}

	method := "DELETE"

	// Test that using an invalid version id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	response, err := deleteVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned deleting version, expected 404, got", response.StatusCode, response.Body)
		return
	}

	// Test that a normal admin user cannot delete a version
	req = events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetAdminUserReqHeader(),
	}
	response, err = deleteVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 403 {
		t.Error("Wrong status code returned, expected 403, got", response.StatusCode, response.Body)
		return
	}

	// Delete deleteme version
	req = events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err = deleteVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	versions, err = getVersionsCheckLength(1)
	if err != nil {
		t.Error(err)
		return
	}

	if versions[0].Number != keepMeVersion.Number {
		t.Errorf("Wrong version remaining. Expected %s, but got %s", versions[0].Number, keepMeVersion.Number)
	}
}

func TestListVersions(t *testing.T) {
	testutils.ResetDb(t)

	version1 := domain.Version{
		Model: gorm.Model{
			ID: 1,
		},
		Number:      "1.1.1",
		Description: "This is version 1",
	}

	version2 := domain.Version{
		Model: gorm.Model{
			ID: 2,
		},
		Number:      "2.2.2",
		Description: "This is version 2",
	}

	userFixtures := []domain.User{
		testutils.AdminUser,
	}

	versionFixtures := []domain.Version{
		version1,
		version2,
	}

	for _, fix := range userFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, fix := range versionFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	_, err := getVersionsCheckLength(2)
	if err != nil {
		t.Error(err)
		return
	}

	method := "GET"

	// List versions
	req := events.APIGatewayProxyRequest{
		HTTPMethod:     method,
		Path:           "/versions",
		PathParameters: map[string]string{},
		Headers:        testutils.GetAdminUserReqHeader(),
	}
	response, err := listVersions(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results := response.Body
	if !strings.Contains(results, version1.Number) || !strings.Contains(results, version2.Number) {
		t.Errorf("listVersions did not include the fixture versions. Got:\n%s\n", results)
	}
}

func TestUpdateVersion(t *testing.T) {
	testutils.ResetDb(t)

	updateMeVersion := domain.Version{
		Model: gorm.Model{
			ID: 1,
		},
		Number:      "6.6.6",
		Description: "This version is to be updated",
	}

	keepMeVersion := domain.Version{
		Model: gorm.Model{
			ID: 2,
		},
		Number:      "3.3.3",
		Description: "This tag is NOT to be updated",
	}

	userFixtures := []domain.User{
		testutils.SuperAdmin,
		testutils.AdminUser,
	}

	versionFixtures := []domain.Version{
		updateMeVersion,
		keepMeVersion,
	}

	for _, fix := range userFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, fix := range versionFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	_, err := getVersionsCheckLength(2)
	if err != nil {
		t.Error(err)
		return
	}

	method := "PUT"

	// Test that using an invalid version id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err := deleteVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned updating version, expected 404, got", response.StatusCode, response.Body)
		return
	}

	// Test that a normal admin user cannot delete a version
	req = events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetAdminUserReqHeader(),
	}
	response, err = updateVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 403 {
		t.Error("Wrong status code returned, expected 403, got", response.StatusCode, response.Body)
		return
	}

	updatedVersion := domain.Version{
		Model: gorm.Model{
			ID: 1,
		},
		Number:      "7.7.7",
		Description: "This version has been updated",
	}
	js, err := json.Marshal(updatedVersion)
	if err != nil {
		t.Error("Unable to marshal update version to JSON, err: ", err.Error())
	}

	// Update version
	req = events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
		Body:    string(js),
	}
	response, err = updateVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	versions, err := getVersionsCheckLength(2)
	if err != nil {
		t.Error(err)
		return
	}

	results := versions[1]

	if results.Number != updatedVersion.Number || results.Description != updatedVersion.Description {
		t.Errorf("Did not update version. \nExpected:\n%+v\n But got:\n%+v", updatedVersion, results)
	}
}

func TestViewVersion(t *testing.T) {
	testutils.ResetDb(t)

	version1 := domain.Version{
		Model: gorm.Model{
			ID: 1,
		},
		Number:      "1.1.1",
		Description: "This is version 1",
	}

	version2 := domain.Version{
		Model: gorm.Model{
			ID: 2,
		},
		Number:      "3.3.3",
		Description: "This is version 2",
	}

	userFixtures := []domain.User{
		testutils.SuperAdmin,
		testutils.AdminUser,
	}

	versionFixtures := []domain.Version{
		version1,
		version2,
	}

	for _, fix := range userFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, fix := range versionFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	_, err := getVersionsCheckLength(2)
	if err != nil {
		t.Error(err)
		return
	}

	method := "GET"

	// Test that using an invalid version id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	response, err := viewVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned viewing version, expected 404, got", response.StatusCode, response.Body)
		return
	}

	// View version1
	req = events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       "/version/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetAdminUserReqHeader(),
	}
	response, err = viewVersion(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results := response.Body
	if !strings.Contains(results, version1.Number) {
		t.Errorf("viewVersion did not include the version1 fixture. Got:\n%s\n", results)
	}
}

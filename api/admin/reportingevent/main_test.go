package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"strings"
	"testing"
)

func TestDeleteEvent(t *testing.T) {
	testutils.ResetDb(t)

	deleteMeEvent := domain.ReportingEvent{
		Date:        "2018-06-25",
		Name:        "Delete Me Test",
		Description: "This event is to be deleted",
	}

	keepMeEvent := domain.ReportingEvent{
		Date:        "2018-06-26",
		Name:        "Keep me",
		Description: "This event is not to be deleted",
	}

	tagFixtures := []*domain.ReportingEvent{&deleteMeEvent, &keepMeEvent}
	for _, fix := range tagFixtures {
		err := db.PutItem(fix)
		if err != nil {
			t.Error(err)
		}
	}

	// Test that using an invalid tag id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/reportingevent/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err := deleteEvent(req)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned deleting tag, expected 404, got", response.StatusCode, response.Body)
	}

	strDeleteID := fmt.Sprintf("%d", deleteMeEvent.ID)

	// Delete deleteme tag and check user and node to ensure they no longer have the tag
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/reportingevent/" + strDeleteID,
		PathParameters: map[string]string{
			"id": strDeleteID,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err = deleteEvent(req)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
	}

	// Check that the event was deleted
	reportingEvents := []domain.ReportingEvent{}
	err = db.ListItems(&reportingEvents, "")
	if err != nil {
		t.Errorf("Error trying to get entries in reporting_events table following the test.\n%s", err.Error())
		return
	}

	if len(reportingEvents) != 1 || reportingEvents[0].ID != keepMeEvent.ID {
		t.Errorf(
			"Wrong reporting_events remaining. Expected 1 with ID %d. \nBut got %d:\n%+v",
			keepMeEvent.ID, len(reportingEvents),
			reportingEvents)
		return
	}
}

func TestListEvents(t *testing.T) {
	testutils.ResetDb(t)
	testutils.CreateAdminUser(t)

	node := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	db.PutItem(&node)

	event1 := domain.ReportingEvent{
		Date:        "2018-06-26",
		Name:        "E1",
		Description: "This is event 1 (nodeless)",
	}

	event2 := domain.ReportingEvent{
		Date:        "2018-06-27",
		Name:        "E2",
		Description: "This is event 2 and has a node",
		NodeID:      node.ID,
	}

	eventFixtures := []domain.ReportingEvent{event1, event2}

	for _, fix := range eventFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	reportingEvents := []domain.ReportingEvent{}
	err := db.ListItems(&reportingEvents, "")
	if err != nil {
		t.Errorf("Error trying to get entries in reporting_events table following the test.\n%s", err.Error())
		return
	}

	method := "GET"

	// List events with superAdmin
	req := events.APIGatewayProxyRequest{
		HTTPMethod:     method,
		Path:           "/reportingevent",
		PathParameters: map[string]string{},
		Headers:        testutils.GetSuperAdminReqHeader(),
	}
	response, err := listEvents(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results := response.Body
	if !strings.Contains(results, event1.Name) || !strings.Contains(results, event2.Name) {
		t.Errorf("listEvents did not include the fixture events. Got:\n%s\n", results)
	}

	// list events with normal admin
	req = events.APIGatewayProxyRequest{
		HTTPMethod:     method,
		Path:           "/reportingevent",
		PathParameters: map[string]string{},
		Headers:        testutils.GetAdminUserReqHeader(),
	}
	response, err = listEvents(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}

	results = response.Body
	if !strings.Contains(results, event1.Name) || strings.Contains(results, event2.Name) {
		t.Errorf("For a normal user, listEvents should have returned event1 only. Got:\n%s\n", results)
	}
}

func updateEventWithSuperAdmin(event domain.ReportingEvent, eventID uint) (events.APIGatewayProxyResponse, string) {
	js, err := json.Marshal(event)
	if err != nil {
		return events.APIGatewayProxyResponse{}, "Unable to marshal update Reporting Event to JSON, err: " + err.Error()
	}

	path := "/reportingevent"
	pathParams := map[string]string{}

	if eventID != 0 {
		strEventID := fmt.Sprintf("%v", eventID)
		path = path + "/" + strEventID
		pathParams["id"] = strEventID
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod:     "PUT",
		Path:           path,
		Headers:        testutils.GetSuperAdminReqHeader(),
		PathParameters: pathParams,
		Body:           string(js),
	}

	resp, err := updateEvent(req)
	if err != nil {
		return resp, "Got error trying to update event, err: " + err.Error()
	}

	return resp, ""
}

func TestUpdateEvent(t *testing.T) {
	testutils.ResetDb(t)

	node := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	db.PutItem(&node)

	updateMeEvent := domain.ReportingEvent{
		NodeID:      node.ID,
		Name:        "Update Me",
		Description: "This event is to be updated",
	}

	keepMeEvent := domain.ReportingEvent{
		Name:        "Keep Me",
		Description: "This event is NOT to be updated",
	}

	eventFixtures := []*domain.ReportingEvent{
		&updateMeEvent,
		&keepMeEvent,
	}

	for _, fix := range eventFixtures {
		err := db.PutItem(fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	resp, errMsg := updateEventWithSuperAdmin(domain.ReportingEvent{}, 404)
	if errMsg != "" {
		t.Error(errMsg)
		return
	}

	if resp.StatusCode != 404 {
		t.Error("Wrong status code returned updating event, expected 404, got", resp.StatusCode, resp.Body)
		return
	}

	newDate := "2018-06-25"
	newDescription := "This event has been updated"

	// Update an existing reporting event
	updateMeEvent.Date = newDate
	updateMeEvent.Description = newDescription

	resp, errMsg = updateEventWithSuperAdmin(updateMeEvent, updateMeEvent.ID)
	if errMsg != "" {
		t.Error(errMsg)
		return
	}

	if resp.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", resp.StatusCode, resp.Body)
		return
	}

	resultEvent := domain.ReportingEvent{}
	db.GetItem(&resultEvent, updateMeEvent.ID)

	if resultEvent.Date != newDate || resultEvent.Description != newDescription {
		t.Errorf(
			"Update did not work. Expected to see Date: %s and Description: %s.\n But got: \n%+v",
			newDate,
			newDescription,
			resultEvent,
		)
		return
	}

	if resultEvent.NodeID != node.ID {
		t.Errorf("Update reporting-event lost the Node ID. \n Got %+v", resultEvent)
	}

}

func TestViewEvent(t *testing.T) {
	testutils.ResetDb(t)

	node := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	db.PutItem(&node)

	event1 := domain.ReportingEvent{
		NodeID: node.ID,
		Name:   "Event1",
	}

	event2 := domain.ReportingEvent{
		Name: "Event2",
	}

	eventFixtures := []*domain.ReportingEvent{
		&event1,
		&event2,
	}

	for _, fix := range eventFixtures {
		err := db.PutItem(fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	strID := fmt.Sprintf("%v", event1.ID)

	// Get event
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/reportingevent/" + strID,
		PathParameters: map[string]string{
			"id": strID,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err := viewEvent(req)
	if err != nil {
		t.Error(err)
		return
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
		return
	}
	if !strings.Contains(response.Body, event1.Name) {
		t.Errorf("Did not get back the event.\nGot: %v", response.Body)
	}
}

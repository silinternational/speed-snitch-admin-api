package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"net/http"
	"testing"
)

func TestDeleteNode(t *testing.T) {
	testutils.ResetDb(t)

	create := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Contacts: []domain.Contact{
			{
				Name:  "test",
				Email: "test@test.com",
			},
		},
	}
	db.PutItem(&create)

	idStr := fmt.Sprintf("%v", create.ID)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/node/" + idStr,
		PathParameters: map[string]string{
			"id": idStr,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := deleteNode(req)
	if err != nil {
		t.Error("Unable to delete node, err: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get 200 response deleting node, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	var node domain.Node
	err = json.Unmarshal([]byte(resp.Body), &node)
	if err != nil {
		t.Error("Unable to unmarshal body into node, err: ", err.Error(), " body: ", resp.Body)
	}

	// try to find node via db to ensure doesn't exist
	var find domain.Node
	err = db.GetItem(&find, create.ID)
	if !gorm.IsRecordNotFoundError(err) {
		t.Error("node still exists after deletion")
	}

	// Check if contact was removed too
	var contact domain.Contact
	err = db.GetItem(&contact, create.Contacts[0].ID)
	if !gorm.IsRecordNotFoundError(err) {
		t.Errorf("contact still exists after node deletion: %+v", contact)
	}

	// try to delete node that doesnt exist
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/node/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err = deleteNode(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Did not get 404 trying to delete node that doesnt exist, got: ", resp.StatusCode)
	}
}

func TestViewNode(t *testing.T) {
	testutils.ResetDb(t)

	create := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			{
				Name:        "test",
				Description: "test",
			},
		},
	}

	db.PutItem(&create)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := viewNode(req)
	if err != nil {
		t.Error("Unable to view node, err: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get 200 response viewing node, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	var node domain.Node
	err = json.Unmarshal([]byte(resp.Body), &node)
	if err != nil {
		t.Error("Unable to unmarshal body into node, err: ", err.Error(), " body: ", resp.Body)
	}

	if node.ID != create.ID {
		t.Errorf("Returned node ID (%v) does not match expected node ID (%v)", node.ID, create.ID)
	}

	// try to view node that doesnt exist
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err = viewNode(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Did not get 404 trying to view node that doesnt exist, got: ", resp.StatusCode)
	}

	// try to view node not authorized to see due to tags not matching
	adminUser := domain.User{
		Role:  domain.UserRoleAdmin,
		Name:  "not super admin",
		Email: "admin@test.com",
		UUID:  "014BF02D-75E6-444B-9231-7BF9C17D42A1",
		Tags: []domain.Tag{
			{
				Name:        "doesnt-match",
				Description: "tag doesn't match",
			},
		},
	}
	db.PutItem(&adminUser)

	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: map[string]string{
			"x-user-uuid": adminUser.UUID,
			"x-user-mail": adminUser.Email,
		},
	}

	resp, err = viewNode(req)
	if err != nil {
		t.Error("Received error trying to view node: ", err.Error())
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Error("Did not get 403 trying to view node that user shouldn't be able to view, got: ", resp.StatusCode, " body: ", resp.Body)
	}

}

func TestListNodes(t *testing.T) {
	testutils.ResetDb(t)

	version := domain.Version{
		Number: "1.0.0",
	}
	db.PutItem(&version)

	tag := domain.Tag{
		Name:        "test",
		Description: "test",
	}
	err := db.PutItem(&tag)
	if err != nil {
		t.Error(err)
		return
	}

	visibleNodes := []domain.Node{
		{
			MacAddr:             "aa:aa:aa:aa:aa:aa",
			RunningVersionID:    version.ID,
			ConfiguredVersionID: version.ID,
			Tags: []domain.Tag{
				tag,
			},
		},
		{
			MacAddr:             "bb:bb:bb:bb:bb:bb",
			RunningVersionID:    version.ID,
			ConfiguredVersionID: version.ID,
			Tags: []domain.Tag{
				tag,
			},
		},
		{
			MacAddr:             "cc:cc:cc:cc:cc:cc",
			RunningVersionID:    version.ID,
			ConfiguredVersionID: version.ID,
			Tags: []domain.Tag{
				tag,
			},
		},
	}

	invisibleNodes := []domain.Node{
		{
			MacAddr:             "dd:dd:dd:dd:dd:dd",
			RunningVersionID:    version.ID,
			ConfiguredVersionID: version.ID,
			Tags: []domain.Tag{
				{
					Name:        "hide1",
					Description: "hide1",
				},
			},
		},
		{
			MacAddr:             "ee:ee:ee:ee:ee:ee",
			RunningVersionID:    version.ID,
			ConfiguredVersionID: version.ID,
			Tags: []domain.Tag{
				{
					Name:        "hide2",
					Description: "hide2",
				},
			},
		},
	}

	for _, i := range visibleNodes {
		err := db.PutItem(&i)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, i := range invisibleNodes {
		err := db.PutItem(&i)
		if err != nil {
			t.Error(err)
			return
		}
	}

	adminUser := domain.User{
		Role:  domain.UserRoleAdmin,
		Name:  "not super admin",
		Email: "admin@test.com",
		UUID:  "014BF02D-75E6-444B-9231-7BF9C17D42A1",
		Tags: []domain.Tag{
			tag,
		},
	}
	err = db.PutItem(&adminUser)
	if err != nil {
		t.Error(err)
		return
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node",
		Headers: map[string]string{
			"x-user-uuid": adminUser.UUID,
			"x-user-mail": adminUser.Email,
		},
	}

	resp, err := listNodes(req)
	if err != nil {
		t.Error("Received error trying to view node: ", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Did not get 200 trying to list nodes, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	var found []domain.Node
	err = json.Unmarshal([]byte(resp.Body), &found)
	if err != nil {
		t.Error("Unable to unmarshal list of nodes, err: ", err.Error(), " body: ", resp.Body)
	}

	if len(found) != len(visibleNodes) {
		t.Error("Did not get back correct number of nodes, expected: ", len(visibleNodes), " got: ", len(found))
	}

	for _, i := range found {
		for _, j := range invisibleNodes {
			if i.ID == j.ID {
				t.Error("Found node in list nodes result that should not have been present, ID: ", i.ID)
			}
		}
	}

}

func TestListNodeTags(t *testing.T) {
	testutils.ResetDb(t)

	version := domain.Version{
		Number: "1.0.0",
	}
	err := db.PutItem(&version)
	if err != nil {
		t.Error(err)
	}

	tag1 := domain.Tag{
		Name:        "test1",
		Description: "test1",
	}

	tag2 := domain.Tag{
		Name:        "test2",
		Description: "test2",
	}

	tag3 := domain.Tag{
		Name:        "test3",
		Description: "test3",
	}

	tag4 := domain.Tag{
		Name:        "test4",
		Description: "test4",
	}

	for _, nextTag := range []*domain.Tag{&tag1, &tag2, &tag3, &tag4} {
		err = db.PutItem(nextTag)
		if err != nil {
			t.Error(err)
			return
		}
	}

	node1 := domain.Node{
		RunningVersionID:    version.ID,
		ConfiguredVersionID: version.ID,
		MacAddr:             "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			tag1,
			tag2,
			tag3,
		},
	}

	node2 := domain.Node{
		MacAddr:             "bb:bb:bb:bb:bb:bb",
		RunningVersionID:    version.ID,
		ConfiguredVersionID: version.ID,
		Tags: []domain.Tag{
			tag1,
			tag4,
		},
	}

	for _, nextNode := range []*domain.Node{&node1, &node2} {
		err = db.PutItem(&nextNode)
		if err != nil {
			t.Error(err)
			return
		}
	}

	node1Id := fmt.Sprintf("%v", node1.ID)
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/" + node1Id + "/tag",
		PathParameters: map[string]string{
			"id": node1Id,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := listNodeTags(req)
	if err != nil {
		t.Error("Unable to list node tags, err: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get 200 response listing node tags, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	var node1Tags []domain.Tag
	err = json.Unmarshal([]byte(resp.Body), &node1Tags)
	if err != nil {
		t.Error("Unable to unmarshal tag list. err: ", err.Error(), " body: ", resp.Body)
	}

	if len(node1Tags) != len(node1.Tags) {
		t.Error("Did not get back right numer of tags for node 1")
	}

	node2Id := fmt.Sprintf("%v", node2.ID)
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/" + node2Id + "/tag",
		PathParameters: map[string]string{
			"id": node2Id,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err = listNodeTags(req)
	if err != nil {
		t.Error("Unable to list node2 tags, err: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get 200 response listing node2 tags, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	var node2Tags []domain.Tag
	err = json.Unmarshal([]byte(resp.Body), &node2Tags)
	if err != nil {
		t.Error("Unable to unmarshal tag list. err: ", err.Error(), " body: ", resp.Body)
	}

	if len(node2Tags) != len(node2.Tags) {
		t.Error("Did not get back right numer of tags for node 2")
	}
}

func TestUpdateNode(t *testing.T) {
	testutils.ResetDb(t)

	version1 := domain.Version{
		Number: "1.0.0",
	}
	err := db.PutItem(&version1)
	if err != nil {
		t.Error(err)
	}

	version2 := domain.Version{
		Number: "2.0.0",
	}
	err = db.PutItem(&version2)
	if err != nil {
		t.Error(err)
	}

	tag1 := domain.Tag{
		Name:        "test1",
		Description: "test1",
	}

	tag2 := domain.Tag{
		Name:        "test2",
		Description: "test2",
	}

	tag3 := domain.Tag{
		Name:        "test3",
		Description: "test3",
	}

	for _, nextTag := range []*domain.Tag{&tag1, &tag2, &tag3} {
		err = db.PutItem(nextTag)
		if err != nil {
			t.Error(err)
			return
		}
	}
	node1 := domain.Node{
		MacAddr:             "aa:aa:aa:aa:aa:aa",
		RunningVersionID:    version1.ID,
		ConfiguredVersionID: version1.ID,
		Contacts:            []domain.Contact{},
		Tags: []domain.Tag{
			tag1,
			tag2,
			tag3,
		},
	}
	err = db.PutItem(&node1)
	if err != nil {
		t.Error(err)
	}

	speedTestNetServer := domain.SpeedTestNetServer{
		Name:        "test stn server",
		ServerID:    "1234",
		Country:     "United States",
		CountryCode: "US",
		Host:        "example.com:8080",
	}
	err = db.PutItem(&speedTestNetServer)
	if err != nil {
		t.Error(err)
	}

	namedServer := domain.NamedServer{
		Name:                 "example",
		Description:          "test example",
		SpeedTestNetServerID: speedTestNetServer.ID,
	}
	err = db.PutItem(&namedServer)
	if err != nil {
		t.Error(err)
	}

	// remove one tag, add tasks, nickname, and notes
	update1 := domain.Node{
		MacAddr:             "aa:aa:aa:aa:aa:aa",
		RunningVersionID:    version1.ID,
		ConfiguredVersionID: version2.ID, // This one should be used
		ConfiguredVersion:   version1,    // This one should be ignored
		Contacts: []domain.Contact{
			{
				Name:  "New Contact",
				Email: "new_contact@test.org",
			},
		},

		Tags: []domain.Tag{
			tag1,
			tag2,
		},
		Nickname: "updated-test",
		Notes:    "created this node via testing",
		Tasks: []domain.Task{
			{
				Type:          domain.TaskTypeSpeedTest,
				NamedServerID: namedServer.ID,
				Schedule:      "* * * * *",
				ServerHost:    namedServer.ServerHost,
			},
		},
	}

	js, err := json.Marshal(update1)
	if err != nil {
		t.Error("Unable to marshal update into json for api call, err: ", err.Error())
	}

	node1Id := fmt.Sprintf("%v", node1.ID)
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/node/" + node1Id,
		PathParameters: map[string]string{
			"id": node1Id,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
		Body:    string(js),
	}

	resp, err := updateNode(req)
	if err != nil {
		t.Error("Unable to update node, err: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get 200 response updating node, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	// fetch node from db to check for updates
	var node domain.Node
	err = db.GetItem(&node, node1.ID)
	if err != nil {
		t.Error("Unable to get node, err: ", err.Error())
		return
	}

	if node.Nickname != update1.Nickname {
		t.Error("Nickname was not updated")
	}

	if len(node.Tags) != len(update1.Tags) {
		t.Error("Tags not updated as expected")
	}

	if len(node.Tasks) != len(update1.Tasks) {
		t.Error("Tasks not updated as expected")
	}

	if len(node.Contacts) != len(update1.Contacts) {
		t.Error("Contacts not updated as expected")
	}

	if node.ConfiguredVersion.Number != version2.Number {
		t.Errorf("Configured Version not updated as expected.")
	}
}

func TestRemoveAssociations(t *testing.T) {
	testutils.ResetDb(t)

	version := domain.Version{
		Number: "1.0.0",
	}
	db.PutItem(&version)

	node1 := domain.Node{
		MacAddr:             "aa:aa:aa:aa:aa:aa",
		Nickname:            "before",
		RunningVersionID:    version.ID,
		ConfiguredVersionID: version.ID,
		Tags: []domain.Tag{
			{
				Name:        "test1",
				Description: "test1",
			},
			{
				Name:        "test2",
				Description: "test2",
			},
			{
				Name:        "test3",
				Description: "test3",
			},
		},
		Contacts: []domain.Contact{
			{
				Name:  "contact 1",
				Email: "contact1@domain.com",
			},
			{
				Name:  "contact 2",
				Email: "contact2@domain.com",
			},
		},
		Tasks: []domain.Task{
			{
				Type: domain.TaskTypePing,
			},
		},
	}
	db.PutItem(&node1)
	node1Id := fmt.Sprintf("%v", node1.ID)

	js := `{"MacAddr": "aa:aa:aa:aa:aa:aa", "Nickname": "after", "Tags": [], "Contacts": [], "Tasks": []}`

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/node/" + node1Id,
		PathParameters: map[string]string{
			"id": node1Id,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
		Body:    string(js),
	}

	resp, err := updateNode(req)
	if err != nil {
		t.Error("Unable to update node, err: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get 200 response updating node, got: ", resp.StatusCode, " body: ", resp.Body)
	}

	// fetch node from db to check for updates
	var node domain.Node
	err = db.GetItem(&node, node1.ID)
	if err != nil {
		t.Error("Unable to get node, err: ", err.Error())
	}

	if node.Nickname != "after" {
		t.Error("Nickname not changed after update")
	}

	if len(node.Tags) != 0 {
		t.Error("Tags still present after update")
	}

	if len(node.Contacts) != 0 {
		t.Error("Contacts still present after update")
	}

	if len(node.Tasks) != 0 {
		t.Error("Tasks still present after update")
	}
}

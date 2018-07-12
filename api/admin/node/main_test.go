package main

import (
	"encoding/json"
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
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	db.PutItem(&create)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/node/1",
		PathParameters: map[string]string{
			"id": "1",
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
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 1,
				},
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
		Model: gorm.Model{
			ID: 2,
		},
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 2,
				},
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

	visibleNodes := []domain.Node{
		{
			Model: gorm.Model{
				ID: 1,
			},
			MacAddr: "aa:aa:aa:aa:aa:aa",
			Tags: []domain.Tag{
				{
					Model: gorm.Model{
						ID: 1,
					},
					Name:        "test",
					Description: "test",
				},
			},
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			MacAddr: "bb:bb:bb:bb:bb:bb",
			Tags: []domain.Tag{
				{
					Model: gorm.Model{
						ID: 1,
					},
					Name:        "test",
					Description: "test",
				},
			},
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			MacAddr: "cc:cc:cc:cc:cc:cc",
			Tags: []domain.Tag{
				{
					Model: gorm.Model{
						ID: 1,
					},
					Name:        "test",
					Description: "test",
				},
			},
		},
	}

	invisibleNodes := []domain.Node{
		{
			Model: gorm.Model{
				ID: 4,
			},
			MacAddr: "dd:dd:dd:dd:dd:dd",
			Tags: []domain.Tag{
				{
					Model: gorm.Model{
						ID: 2,
					},
					Name:        "hide",
					Description: "hide",
				},
			},
		},
		{
			Model: gorm.Model{
				ID: 5,
			},
			MacAddr: "ee:ee:ee:ee:ee:ee",
			Tags: []domain.Tag{
				{
					Model: gorm.Model{
						ID: 2,
					},
					Name:        "hide",
					Description: "hide",
				},
			},
		},
	}

	for _, i := range visibleNodes {
		db.PutItem(&i)
	}

	for _, i := range invisibleNodes {
		db.PutItem(&i)
	}

	adminUser := domain.User{
		Role:  domain.UserRoleAdmin,
		Name:  "not super admin",
		Email: "admin@test.com",
		UUID:  "014BF02D-75E6-444B-9231-7BF9C17D42A1",
		Model: gorm.Model{
			ID: 2,
		},
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 1,
				},
				Name:        "test",
				Description: "test",
			},
		},
	}
	db.PutItem(&adminUser)

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
		t.Error("Did not get 200 trying tolist nodes, got: ", resp.StatusCode, " body: ", resp.Body)
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

	node1 := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 1,
				},
				Name:        "test1",
				Description: "test1",
			},
			{
				Model: gorm.Model{
					ID: 2,
				},
				Name:        "test2",
				Description: "test2",
			},
			{
				Model: gorm.Model{
					ID: 3,
				},
				Name:        "test3",
				Description: "test3",
			},
		},
	}

	node2 := domain.Node{
		Model: gorm.Model{
			ID: 2,
		},
		MacAddr: "bb:bb:bb:bb:bb:bb",
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 1,
				},
				Name:        "test1",
				Description: "test1",
			},
			{
				Model: gorm.Model{
					ID: 4,
				},
				Name:        "test4",
				Description: "test4",
			},
		},
	}

	db.PutItem(&node1)
	db.PutItem(&node2)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/1/tag",
		PathParameters: map[string]string{
			"id": "1",
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

	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/node/2/tag",
		PathParameters: map[string]string{
			"id": "2",
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

	node1 := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 1,
				},
				Name:        "test1",
				Description: "test1",
			},
			{
				Model: gorm.Model{
					ID: 2,
				},
				Name:        "test2",
				Description: "test2",
			},
			{
				Model: gorm.Model{
					ID: 3,
				},
				Name:        "test3",
				Description: "test3",
			},
		},
	}
	db.PutItem(&node1)

	speedTestNetServer := domain.SpeedTestNetServer{
		Model: gorm.Model{
			ID: 1,
		},
		Name:        "test stn server",
		ServerID:    "1234",
		Country:     "United States",
		CountryCode: "US",
		Host:        "example.com:8080",
	}
	db.PutItem(&speedTestNetServer)

	namedServer := domain.NamedServer{
		Model: gorm.Model{
			ID: 1,
		},
		Name:                 "example",
		Description:          "test example",
		SpeedTestNetServerID: speedTestNetServer.ID,
	}
	db.PutItem(&namedServer)

	// remove one tag, add tasks, nickname, and notes
	update1 := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			{
				Model: gorm.Model{
					ID: 1,
				},
				Name:        "test1",
				Description: "test1",
			},
			{
				Model: gorm.Model{
					ID: 2,
				},
				Name:        "test2",
				Description: "test2",
			},
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

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/node/1",
		PathParameters: map[string]string{
			"id": "1",
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

	if node.Nickname != update1.Nickname {
		t.Error("Nickname was not updated")
	}
}

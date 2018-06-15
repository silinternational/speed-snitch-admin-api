package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"testing"
)

func TestDeleteTag(t *testing.T) {
	db.FlushTables(t)

	deleteMeTag := domain.Tag{
		ID:          "tag-deleteme",
		UID:         "deleteme",
		Name:        "Delete Me Test",
		Description: "This tag is to be deleted",
	}

	keepMeTag := domain.Tag{
		ID:          "tag-keepme",
		UID:         "keepme",
		Name:        "Keep me",
		Description: "This tag is not to be deleted",
	}

	nodeFixtures := []domain.Node{
		{
			ID:      "node-aa:aa:aa:aa:aa:aa",
			MacAddr: "aa:aa:aa:aa:aa:aa",
			Tags: []domain.Tag{
				deleteMeTag,
				keepMeTag,
			},
		},
	}

	userFixtures := []domain.User{
		{
			ID:     "user-super",
			UID:    "super",
			UserID: "super_admin",
			Role:   domain.UserRoleSuperAdmin,
		},
		{
			ID:     "user-changed",
			UID:    "changed",
			UserID: "changed_test",
			Role:   domain.UserRoleAdmin,
			Tags: []domain.Tag{
				deleteMeTag,
				keepMeTag,
			},
		},
		{
			ID:     "user-notchanged",
			UID:    "notchanged",
			UserID: "notchanged_test",
			Role:   domain.UserRoleAdmin,
			Tags: []domain.Tag{
				keepMeTag,
			},
		},
	}

	tagFixtures := []domain.Tag{
		deleteMeTag,
		keepMeTag,
	}

	for _, fix := range nodeFixtures {
		err := db.PutItem(domain.DataTable, fix)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	for _, fix := range userFixtures {
		err := db.PutItem(domain.DataTable, fix)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	for _, fix := range tagFixtures {
		err := db.PutItem(domain.DataTable, fix)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	// Test that using an invalid tag uid results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/tag/does-not-exist",
		PathParameters: map[string]string{
			"uid": "does-not-exist",
		},
		Headers: map[string]string{
			"x-user-id": "super_admin",
		},
	}
	response, err := deleteTag(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned, expected 404, got", response.StatusCode, response.Body)
	}

	// Delete deleteme tag and check user and node to ensure they no longer have the tag
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/tag/deleteme",
		PathParameters: map[string]string{
			"uid": "deleteme",
		},
		Headers: map[string]string{
			"x-user-id": "super_admin",
		},
	}
	response, err = deleteTag(req)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
	}

	user, _ := db.GetUser("changed")
	hasTag, _ := domain.InArray(deleteMeTag, user.Tags)
	if hasTag {
		t.Errorf("Tag deleteme still present in user tags after deletion. User tags: %v", user.Tags)
		t.Fail()
	}
	if len(user.Tags) != 1 {
		t.Errorf("User does not have one tag (keepme) as expected, has tags: %v", user.Tags)
	}

	node, _ := db.GetNode("aa:aa:aa:aa:aa:aa")
	hasTag, _ = domain.InArray(deleteMeTag, node.Tags)
	if hasTag {
		t.Errorf("Tag deleteme still present in node tags after deletion. User tags: %v", node.Tags)
		t.Fail()
	}
	if len(node.Tags) != 1 {
		t.Error("Node does not have one tag (keepme) as expected, has tags:", node.Tags)
	}

	// Get other user who should not have been changed to ensure they were not
	notChangedUser, err := db.GetUser("notchanged")
	if len(notChangedUser.Tags) != 1 {
		t.Errorf("User does not have one tag (notchanged) as expected, has tags: %v", notChangedUser.Tags)
	}

}

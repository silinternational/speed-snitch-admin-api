package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"testing"
)

func TestDeleteTag(t *testing.T) {
	db.FlushTables(t)

	nodeFixtures := []domain.Node{
		{
			ID:      "node-aa:aa:aa:aa:aa:aa",
			MacAddr: "aa:aa:aa:aa:aa:aa",
			TagUIDs: []string{"deleteme", "keepme"},
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
			ID:      "user-changed",
			UID:     "changed",
			UserID:  "changed_test",
			Role:    domain.UserRoleAdmin,
			TagUIDs: []string{"deleteme", "keepme"},
		},
		{
			ID:      "user-notchanged",
			UID:     "notchanged",
			UserID:  "notchanged_test",
			Role:    domain.UserRoleAdmin,
			TagUIDs: []string{"notchanged"},
		},
	}

	tagFixtures := []domain.Tag{
		{
			ID:          "tag-deleteme",
			UID:         "deleteme",
			Name:        "Delete Me Test",
			Description: "This tag is to be deleted",
		},
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

	var user domain.User
	err = db.GetItem(domain.DataTable, domain.DataTypeUser, "changed", &user)
	hasTag, _ := domain.InArray("deleteme", user.TagUIDs)
	if hasTag {
		t.Error("Tag deleteme still present in user tags after deletion. User tags: ", user.TagUIDs)
		t.Fail()
	}
	if len(user.TagUIDs) != 1 {
		t.Error("User does not have one tag (keepme) as expected, has tags:", user.TagUIDs)
	}

	var node domain.Node
	err = db.GetItem(domain.DataTable, domain.DataTypeNode, "aa:aa:aa:aa:aa:aa", &node)
	hasTag, _ = domain.InArray("deleteme", node.TagUIDs)
	if hasTag {
		t.Error("Tag deleteme still present in node tags after deletion. User tags: ", node.TagUIDs)
		t.Fail()
	}
	if len(node.TagUIDs) != 1 {
		t.Error("Node does not have one tag (keepme) as expected, has tags:", node.TagUIDs)
	}

	// Get other user who should not have been changed to ensure they were not
	var notChangedUser domain.User
	err = db.GetItem(domain.DataTable, domain.DataTypeUser, "notchanged", &notChangedUser)
	if len(notChangedUser.TagUIDs) != 1 {
		t.Error("User does not have one tag (notchanged) as expected, has tags:", notChangedUser.TagUIDs)
	}

}

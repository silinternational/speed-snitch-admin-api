package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"testing"
)

func TestDeleteTag(t *testing.T) {
	testutils.ResetDb(t)

	deleteMeTag := domain.Tag{
		Model: gorm.Model{
			ID: 1,
		},
		Name:        "Delete Me Test",
		Description: "This tag is to be deleted",
	}

	keepMeTag := domain.Tag{
		Model: gorm.Model{
			ID: 2,
		},
		Name:        "Keep me",
		Description: "This tag is not to be deleted",
	}

	tagFixtures := []domain.Tag{deleteMeTag, keepMeTag}
	for _, fix := range tagFixtures {
		err := db.PutItem(&fix)
		if err != nil {
			t.Error(err)
		}
	}

	changedNode := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
		Tags: []domain.Tag{
			deleteMeTag,
			keepMeTag,
		},
	}

	// Create the node in the database
	err := db.PutItemWithAssociations(
		&changedNode,
		[]domain.AssociationReplacement{
			{Replacement: deleteMeTag, AssociationName: "nodes"},
			{Replacement: keepMeTag, AssociationName: "nodes"},
		},
	)

	changedUser := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "changed_test@mail.com",
		Role:  domain.UserRoleAdmin,
	}

	notChangedUser := domain.User{
		Model: gorm.Model{
			ID: 3,
		},
		UUID:  "3",
		Email: "notchanged_test@mail.com",
		Role:  domain.UserRoleAdmin,
	}

	// Create the user in the database
	err = db.PutItemWithAssociations(
		&changedUser,
		[]domain.AssociationReplacement{
			{Replacement: deleteMeTag, AssociationName: "tags"},
			{Replacement: keepMeTag, AssociationName: "tags"},
		},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	err = db.PutItemWithAssociations(
		&notChangedUser,
		[]domain.AssociationReplacement{{Replacement: keepMeTag, AssociationName: "tags"}},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	// Test that using an invalid tag id results in 404 error
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/tag/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err := deleteTag(req)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != 404 {
		t.Error("Wrong status code returned deleting tag, expected 404, got", response.StatusCode, response.Body)
	}

	// Delete deleteme tag and check user and node to ensure they no longer have the tag
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/tag/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}
	response, err = deleteTag(req)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != 200 {
		t.Error("Wrong status code returned, expected 200, got", response.StatusCode, response.Body)
	}

	var user domain.User
	err = db.GetItem(&user, changedUser.ID)
	if err != nil {
		t.Error("Got error trying to get changed user from db: ", err)
	}
	hasTag, _ := domain.InArray(deleteMeTag, user.Tags)
	if hasTag {
		t.Errorf("Tag deleteme still present in user tags after deletion. User tags: %v", user.Tags)
	}
	if len(user.Tags) != 1 {
		t.Errorf("User does not have one tag (keepme) as expected, has tags: %v", user.Tags)
	}

	var node domain.Node
	err = db.GetItem(&node, changedNode.ID)
	if err != nil {
		t.Error("Got error trying to get changed node from db: ", err)
	}
	hasTag, _ = domain.InArray(deleteMeTag, node.Tags)
	if hasTag {
		t.Errorf("Tag deleteme still present in node tags after deletion. User tags: %v", node.Tags)
	}
	if len(node.Tags) != 1 {
		t.Errorf("Node does not have one tag (keepme) as expected, has tags: %v", node.Tags)
	}

	// Get other user who should not have been changed to ensure they were not
	var unchangedUser domain.User
	err = db.GetItem(&unchangedUser, notChangedUser.ID)
	if err != nil {
		t.Error("Got error trying to get unchanged user from db: ", err)
	}
	if len(notChangedUser.Tags) != 1 {
		t.Errorf("User does not have one tag (notchanged) as expected, has tags: %v", notChangedUser.Tags)
	}

}

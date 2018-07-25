package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"testing"
)

func TestDeleteTag(t *testing.T) {
	testutils.ResetDb(t)

	deleteMeTag := domain.Tag{
		Name:        "Delete Me Test",
		Description: "This tag is to be deleted",
	}

	keepMeTag := domain.Tag{
		Name:        "Keep me",
		Description: "This tag is not to be deleted",
	}

	tagFixtures := []*domain.Tag{&deleteMeTag, &keepMeTag}
	for _, fix := range tagFixtures {
		err := db.PutItem(fix)
		if err != nil {
			t.Error(err)
		}
	}

	changedNode := domain.Node{
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}

	// Create the node in the database
	err := db.PutItemWithAssociations(
		&changedNode,
		[]domain.AssociationReplacements{
			{AssociationName: "Tags", Replacements: []domain.Tag{deleteMeTag, keepMeTag}},
		},
	)

	var nodeTags []domain.NodeTags
	err = db.ListItems(&nodeTags, "")
	if err != nil {
		t.Errorf("Error trying to get entries in node_tags table before the test.\n%s", err.Error())
		return
	}

	if len(nodeTags) != 2 {
		t.Errorf("Wrong number of node_tags saved with fixtures. Expected: 2. But got: %d", len(nodeTags))
		return
	}

	changedUser := domain.User{
		UUID:  "2",
		Email: "changed_test@mail.com",
		Role:  domain.UserRoleAdmin,
	}

	notChangedUser := domain.User{
		UUID:  "3",
		Email: "notchanged_test@mail.com",
		Role:  domain.UserRoleAdmin,
	}

	// Create the user in the database
	err = db.PutItemWithAssociations(
		&changedUser,
		[]domain.AssociationReplacements{
			{Replacements: []domain.Tag{deleteMeTag, keepMeTag}, AssociationName: "tags"},
		},
	)

	if err != nil {
		t.Error("Got Error loading user fixture.\n", err.Error())
		return
	}

	err = db.PutItemWithAssociations(
		&notChangedUser,
		[]domain.AssociationReplacements{{Replacements: []domain.Tag{keepMeTag}, AssociationName: "tags"}},
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

	strDeleteID := fmt.Sprintf("%d", deleteMeTag.ID)

	// Delete deleteme tag and check user and node to ensure they no longer have the tag
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/tag/" + strDeleteID,
		PathParameters: map[string]string{
			"id": strDeleteID,
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

	// Check that the node_tag was deleted
	nodeTags = []domain.NodeTags{}
	err = db.ListItems(&nodeTags, "")
	if err != nil {
		t.Errorf("Error trying to get entries in node_tags table following the test.\n%s", err.Error())
		return
	}

	if len(nodeTags) != 1 || nodeTags[0].TagID != keepMeTag.ID {
		t.Errorf("Wrong node_tags remaining. Expected 1 with ID %d. \nBut got %d:\n%+v", keepMeTag.ID, len(nodeTags), nodeTags)
		return
	}

	// Check that the user_tags were deleted (leaving one per user)
	userTags := []domain.UserTags{}
	err = db.ListItems(&userTags, "")
	if err != nil {
		t.Errorf("Error trying to get entries in user_tags table following the test.\n%s", err.Error())
		return
	}

	if len(userTags) != 2 || userTags[0].TagID != keepMeTag.ID || userTags[1].TagID != keepMeTag.ID {
		t.Errorf("Wrong user_tags remaining. Expected 1 with ID %d. \nBut got %d:\n%+v", keepMeTag.ID, len(userTags), userTags)
		return
	}
}

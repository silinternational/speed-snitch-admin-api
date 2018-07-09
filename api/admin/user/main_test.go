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

func TestDeleteUser(t *testing.T) {
	testutils.ResetDb(t)

	createUser := domain.User{
		Model: gorm.Model{
			ID: 1,
		},
		Name:  "test",
		Email: "test@test.com",
		UUID:  "abc123",
	}

	err := db.PutItem(&createUser)
	if err != nil {
		t.Error("Error creating test user: ", err.Error())
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/user/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := deleteUser(req)

	if err != nil {
		t.Error("Error deleting user: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get expected 200 response for deleting uesr, got: ", resp.StatusCode, " body: ", string(resp.Body))
	}

	// query db directly to make sure user no longer exists
	var findUser domain.User
	err = db.GetItem(&findUser, "1")
	if !gorm.IsRecordNotFoundError(err) {
		t.Error("Did not get a not found error as expected")
	}
}

func TestViewMe(t *testing.T) {
	testutils.ResetDb(t)

	user1 := domain.User{
		Model: gorm.Model{
			ID: 1,
		},
		UUID:  "1",
		Email: "user1@test.com",
	}

	user2 := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "user2@test.com",
	}

	db.PutItem(&user1)
	db.PutItem(&user2)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/me",
		Headers: map[string]string{
			"x-user-uuid": "1",
			"x-user-mail": "user1@test.com",
		},
	}

	resp, err := viewMe(req)
	if err != nil {
		t.Error("Error calling view me: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get a 200 response for view me, got: ", resp.StatusCode)
	}

	var respUser domain.User
	err = json.Unmarshal([]byte(resp.Body), &respUser)
	if err != nil {
		t.Error("Unable to unmarshal response into user, err: ", err.Error(), " body: ", resp.Body)
	}

	if respUser.Model.ID != user1.Model.ID {
		t.Error("Returned user does not match expected user")
	}

	// do it again with second user
	req2 := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/me",
		Headers: map[string]string{
			"x-user-uuid": "2",
			"x-user-mail": "user2@test.com",
		},
	}

	resp2, err := viewMe(req2)
	if err != nil {
		t.Error("Error calling view me: ", err.Error())
	}

	var respUser2 domain.User
	err = json.Unmarshal([]byte(resp2.Body), &respUser2)
	if err != nil {
		t.Error("Unable to unmarshal response into user, err: ", err.Error(), " body: ", resp2.Body)
	}

	if respUser2.Model.ID != user2.Model.ID {
		t.Errorf("Returned user does not match expected user. Expected: %+v, got %+v", user2, respUser2)
	}

	//test missing uuid or email
	req3 := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/me",
		Headers: map[string]string{
			"x-user-uuid": "2",
		},
	}
	resp3, err := viewMe(req3)
	if resp3.StatusCode == http.StatusOK {
		t.Error("Should have gotten an error when missing x-user-uuid header, resp body: ", resp3.Body)
	}

	req4 := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/me",
		Headers: map[string]string{
			"x-user-mail": "user2@test.com",
		},
	}
	resp4, err := viewMe(req4)
	if resp4.StatusCode == http.StatusOK {
		t.Error("Should have gotten an error when missing x-user-mail header")
	}
}

func TestViewUser(t *testing.T) {
	testutils.ResetDb(t)

	user1 := domain.User{
		Model: gorm.Model{
			ID: 1,
		},
		UUID:  "1",
		Email: "user1@test.com",
	}

	user2 := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "user2@test.com",
	}

	db.PutItem(&user1)
	db.PutItem(&user2)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/1",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := viewUser(req)
	if err != nil {
		t.Error("Error calling view user: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get a 200 response for view user, got: ", resp.StatusCode)
	}

	var user domain.User
	err = json.Unmarshal([]byte(resp.Body), &user)
	if err != nil {
		t.Error("Unable to unmarshal body into user, err: ", err.Error(), " body: ", resp.Body)
	}

	if user.UUID != user1.UUID {
		t.Errorf("Did not get correct user back, expected UUID %s, got UUID %s", user1.UUID, user.UUID)
	}

	// try again for invalid user
	req = events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/404",
		PathParameters: map[string]string{
			"id": "404",
		},
		Headers: map[string]string{
			"x-user-uuid": fmt.Sprintf("%v", user1.Model.ID),
			"x-user-mail": user1.Email,
		},
	}

	resp, err = viewUser(req)
	if err != nil {
		t.Error("Error calling view user: ", err.Error())
	}

	if resp.StatusCode != 404 {
		t.Error("Did not get a 404 response for invalid view user, got: ", resp.StatusCode)
	}
}

func TestListUserTags(t *testing.T) {
	testutils.ResetDb(t)

	tags := []domain.Tag{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "one",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Name: "two",
		},
	}

	for _, tag := range tags {
		err := db.PutItem(&tag)
		if err != nil {
			t.Error("Unable to put tag for testing, tag: ", tag.Name, " err: ", err.Error())
		}
	}

	user1 := domain.User{
		Model: gorm.Model{
			ID: 1,
		},
		UUID:  "1",
		Email: "user1@test.com",
		Tags:  tags,
	}

	user2 := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "user2@test.com",
		Tags:  tags,
	}

	db.PutItem(&user1)
	db.PutItem(&user2)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/1/tag",
		PathParameters: map[string]string{
			"id": "1",
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := listUserTags(req)
	if err != nil {
		t.Error("Error calling list user tags: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get a 200 response for list user tags, got: ", resp.StatusCode)
	}

	var foundTags []domain.Tag
	err = json.Unmarshal([]byte(resp.Body), &foundTags)
	if err != nil {
		t.Error("Unable to unmarshal body into list of tags. Err: ", err.Error(), " body: ", resp.Body)
	}

	if len(foundTags) != len(user1.Tags) {
		t.Error("Did not get back same number of tags as expected. Got: ", len(foundTags))
	}
}

func TestListUsers(t *testing.T) {
	testutils.ResetDb(t)

	user1 := domain.User{
		Model: gorm.Model{
			ID: 1,
		},
		UUID:  "1",
		Email: "user1@test.com",
	}

	user2 := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "user2@test.com",
	}

	db.PutItem(&user1)
	db.PutItem(&user2)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user",
		Headers:    testutils.GetSuperAdminReqHeader(),
	}

	resp, err := listUsers(req)
	if err != nil {
		t.Error("Error calling list users: ", err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get a 200 response for list users, got: ", resp.StatusCode)
	}

	var users []domain.User
	err = json.Unmarshal([]byte(resp.Body), &users)
	if err != nil {
		t.Error("Unable to unmarshal json into array of users. Err: ", err.Error(), " body: ", resp.Body)
	}

	if len(users) != 2 {
		t.Error("Did not get back number of users expected in list users call, got: ", len(users), " expected: 2")
	}
}

func TestUpdateUser(t *testing.T) {
	testutils.ResetDb(t)

	createUser := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "user2@test.com",
		Role:  domain.UserRoleAdmin,
	}

	updatedUser := domain.User{
		Model: gorm.Model{
			ID: 2,
		},
		UUID:  "2",
		Email: "updated@test.com",
		Role:  domain.UserRoleAdmin,
	}
	js, err := json.Marshal(updatedUser)
	if err != nil {
		t.Error("Unable to marshal update user to JSON, err: ", err.Error())
	}

	db.PutItem(createUser)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/user/2",
		Headers:    testutils.GetSuperAdminReqHeader(),
		PathParameters: map[string]string{
			"id": "2",
		},
		Body: string(js),
	}

	resp, err := updateUser(req)
	if err != nil {
		t.Error("Got error trying to update user, err: ", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("Got back wrong status code updating user, got: ", resp.StatusCode)
	}

	var respUser domain.User
	err = json.Unmarshal([]byte(resp.Body), &respUser)
	if err != nil {
		t.Error("Unable to unmarshal body into updated user. err: ", err.Error(), " body: ", resp.Body)
	}

	if respUser.Email != updatedUser.Email {
		t.Errorf("Updated user's email does not match what is should. Got: %s, expected: %s", respUser.Email, updatedUser.Email)
	}

}

func TestIsRoleValid(t *testing.T) {
	tests := []struct {
		Role    string
		IsValid bool
	}{
		{
			Role:    domain.UserRoleSuperAdmin,
			IsValid: true,
		},
		{
			Role:    domain.UserRoleAdmin,
			IsValid: true,
		},
		{
			Role:    "frog",
			IsValid: false,
		},
	}

	for _, i := range tests {
		isValid := isValidRole(i.Role)
		if isValid != i.IsValid {
			t.Errorf("Role (%s) failed validation check. Should have been %v, got %v", i.Role, i.IsValid, isValid)
		}
	}
}

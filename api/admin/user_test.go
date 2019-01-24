package admin

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

func updateUserWithSuperAdmin(testUser domain.User, userID uint) (events.APIGatewayProxyResponse, string) {
	js, err := json.Marshal(testUser)
	if err != nil {
		return events.APIGatewayProxyResponse{}, "Unable to marshal update User to JSON, err: " + err.Error()
	}

	path := "/user"
	pathParams := map[string]string{}

	if userID != 0 {
		strUserID := fmt.Sprintf("%v", userID)
		path = path + "/" + strUserID
		pathParams["id"] = strUserID
	}

	req := events.APIGatewayProxyRequest{
		HTTPMethod:     "PUT",
		Path:           path,
		Headers:        testutils.GetSuperAdminReqHeader(),
		PathParameters: pathParams,
		Body:           string(js),
	}

	resp, err := updateUser(req)
	if err != nil {
		return resp, "Got error trying to update user, err: " + err.Error()
	}

	return resp, ""
}

func TestDeleteUser(t *testing.T) {
	testutils.ResetDb(t)

	targetUserTag1 := domain.Tag{
		Name:        "TargetUserTag1",
		Description: "First tag for the target user",
	}

	otherTag := domain.Tag{
		Name:        "otherTag",
		Description: "This tag is not for the target user",
	}

	targetUserTag2 := domain.Tag{
		Name:        "TargetUserTag2",
		Description: "Second tag for the target user",
	}

	tagFixtures := []*domain.Tag{&targetUserTag1, &otherTag, &targetUserTag2}
	for _, fix := range tagFixtures {
		err := db.PutItem(fix)
		if err != nil {
			t.Error(err)
			return
		}
	}

	testUser := domain.User{
		Name:  "test",
		Email: "test@test.com",
		UUID:  "abc123",
	}

	// Save the user in the database
	err := db.PutItemWithAssociations(
		&testUser,
		[]domain.AssociationReplacements{
			{AssociationName: "Tags", Replacements: []domain.Tag{targetUserTag1, targetUserTag2}},
		},
	)
	if err != nil {
		t.Error("Error creating test user: ", err.Error())
		return
	}

	// Check that the users got loaded
	users := []domain.User{}
	err = db.ListItems(&users, "id asc")
	if err != nil {
		t.Errorf("Error calling list users: %s", err.Error())
		return
	}

	// Including the SuperAdmin
	if len(users) != 2 {
		t.Errorf("Wrong number of user fixtures loaded. Expected: 2. But got: %d", len(users))
		return
	}

	var userTags []domain.UserTags
	err = db.ListItems(&userTags, "")
	if err != nil {
		t.Errorf("Error trying to get entries in user_tags table before the test.\n%s", err.Error())
		return
	}

	if len(userTags) != 2 {
		t.Errorf("Wrong number of user_tags saved with fixtures. Expected: 2. But got: %d", len(userTags))
		return
	}

	userID := fmt.Sprintf("%d", testUser.ID)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/user/" + userID,
		PathParameters: map[string]string{
			"id": userID,
		},
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err := deleteUser(req)

	if err != nil {
		t.Error("Error deleting user: ", err.Error())
		return
	}

	if resp.StatusCode != 200 {
		t.Error("Did not get expected 200 response for deleting user, got: ", resp.StatusCode, " body: ", string(resp.Body))
		//return
	}

	// query db directly to make sure user no longer exists
	var findUser domain.User
	err = db.GetItem(&findUser, testUser.ID)
	if !gorm.IsRecordNotFoundError(err) {
		t.Error("Did not get a not found error as expected")
	}

	userTags = []domain.UserTags{}
	err = db.ListItems(&userTags, "")
	if err != nil {
		t.Errorf("Error trying to get entries in user_tags table following the test.\n%s", err.Error())
		return
	}

	if len(userTags) != 0 {
		t.Errorf("Wrong number of user_tags remaining. Expected: 0. But got: %d", len(userTags))
		return
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

	users := []domain.User{}
	err := db.ListItems(&users, "id asc")
	if err != nil {
		t.Errorf("Error calling list users: %s", err.Error())
		return
	}

	if len(users) != 2 {
		t.Errorf("Wrong number of user fixtures loaded. Expected: 2. But got: %d", len(users))
	}

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

	if respUser.ID != user1.ID {
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

	if respUser2.ID != user2.ID {
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
			ID: 11,
		},
		UUID:  "11",
		Email: "user1@test.com",
	}

	user2 := domain.User{
		Model: gorm.Model{
			ID: 12,
		},
		UUID:  "12",
		Email: "user2@test.com",
	}

	db.PutItem(&user1)
	db.PutItem(&user2)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/user/11",
		PathParameters: map[string]string{
			"id": "11",
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
		Headers: testutils.GetSuperAdminReqHeader(),
	}

	resp, err = viewUser(req)
	if err != nil {
		t.Error("Error calling view user: ", err.Error())
	}

	if resp.StatusCode != 404 {
		t.Error("Did not get a 404 response for invalid view user, got: ", resp.StatusCode)
	}
}

func TestListUsers(t *testing.T) {
	testutils.ResetDb(t)

	user1 := domain.User{
		Model: gorm.Model{
			ID: 11,
		},
		UUID:  "11",
		Email: "user1@test.com",
	}

	user2 := domain.User{
		Model: gorm.Model{
			ID: 12,
		},
		UUID:  "12",
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

	// Includes SuperAdmin
	if len(users) != 3 {
		t.Error("Did not get back number of users expected in list users call, got: ", len(users), " expected: 3")
	}
}

func TestUpdateUser(t *testing.T) {
	testutils.ResetDb(t)

	createUser := domain.User{
		Model: gorm.Model{
			ID: 11,
		},
		UUID:  "11",
		Email: "user2@test.com",
		Role:  domain.UserRoleAdmin,
	}
	db.PutItem(createUser)

	updatedUser := domain.User{
		UUID:  "22",
		Email: "old22@test.com",
		Role:  domain.UserRoleAdmin,
	}
	resp, errMsg := updateUserWithSuperAdmin(updatedUser, 0)
	if errMsg != "" {
		t.Error(errMsg)
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("Got back wrong status code updating user, got: ", resp.StatusCode, resp.Body)
		return
	}

	// Update existing User
	updatedUser.Email = "new22@test.com"
	resp, errMsg = updateUserWithSuperAdmin(updatedUser, updatedUser.ID)
	if errMsg != "" {
		t.Error(errMsg)
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("Got back wrong status code updating user, got: ", resp.StatusCode, resp.Body)
		return
	}

	// Check the new email for updated User
	var respUser domain.User
	err := json.Unmarshal([]byte(resp.Body), &respUser)
	if err != nil {
		t.Error("Unable to unmarshal body into updated user. err: ", err.Error(), " body: ", resp.Body)
		return
	}

	if respUser.Email != updatedUser.Email {
		t.Errorf("Updated user's email does not match what it should. Got: %s, expected: %s", respUser.Email, updatedUser.Email)
		return
	}

	// Check for a 409 when creating a new User with an email that already has been used
	newUser := domain.User{
		UUID:  "333",
		Email: createUser.Email,
		Role:  domain.UserRoleAdmin,
	}

	resp, errMsg = updateUserWithSuperAdmin(newUser, 0)
	if err != nil {
		t.Error("Got error trying to update user, err: ", err.Error())
		return
	}

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Got back wrong status code updating user. Expected %v, but got: %v", http.StatusConflict, resp.StatusCode)
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

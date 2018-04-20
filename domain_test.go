package domain

import (
	"testing"
)

type testMACAddr struct {
	MA1         string
	MA2         string
	ExpectError bool
}

func TestCleanMACAddress(t *testing.T) {

	testMAs := []testMACAddr{
		{
			MA1:         "112233AAbbCC",
			MA2:         "112233aabbcc",
			ExpectError: false,
		},
		{
			MA1:         "11:22:33:AA:bb:CC",
			MA2:         "11:22:33:aa:bb:cc",
			ExpectError: false,
		},
		{
			MA1:         "11:22-33:AA-bb:CC",
			MA2:         "11:22-33:aa-bb:cc",
			ExpectError: false,
		},
		{
			MA1:         "112233AAbb",
			MA2:         "",
			ExpectError: true, // too short
		},
		{
			MA1:         "112233AAbbCCDD",
			MA2:         "",
			ExpectError: true, // too long
		},
		{
			MA1:         "G12233AAbbCC",
			MA2:         "",
			ExpectError: true, // bad letter
		},
		{
			MA1:         "11:22-33aabbcc",
			MA2:         "",
			ExpectError: true, // not enough delimiters
		},
	}

	for _, ma := range testMAs {

		resultMA, err := CleanMACAddress(ma.MA1)
		if ma.ExpectError && err == nil {
			t.Errorf("For test MAC Address %s expected an error but did not get one.", ma.MA1)
		} else if !ma.ExpectError && err != nil {
			t.Errorf("For test MAC Address %s did not expect an error but got one:\n\t%s.", ma.MA1, err.Error())
		}

		if resultMA != ma.MA2 {
			t.Errorf(`For test MAC Address %s expected: "%s", but got "%s".`, ma.MA1, ma.MA2, resultMA)
		}
	}
}

func TestGetFieldNamesFromJson(t *testing.T) {
	versionNumber := "0.0.1"
	newDescription := "New Version"

	oldObj := Version{
		versionNumber,
		"Old Version",
	}

	newJson := `{
	"Number": "` + versionNumber + `",
	"Description": "` + newDescription + `",
	"Other": "stuff"
}`

	jsonTags := GetFieldNamesFromJson(oldObj, newJson)

	expected := 2
	results := len(jsonTags)

	if expected != results {
		t.Errorf("Wrong number of tags found. Expected %d, but got %d", expected, results)
		return
	}

	expectedTags := []string{"Number", "Description"}

	for index, tag := range expectedTags {
		results := jsonTags[index]
		if tag != results {
			t.Errorf("Wrong tag found at index %d. Expected %s, but got %s", index, tag, results)
			return
		}
	}
}

type UserData struct {
	ID    string
	Name  string
	Email string
	Role  string
	Tags  []Tag
	User  User
}

func GetAdminUserData() UserData {
	// Sending original copies of the values to ensure that changes
	// to the original User don't mask bad results
	return UserData{
		"123",
		"Andy Admin",
		"andy_admin@some.org",
		"Admin",
		[]Tag{
			{"Africa", "Locations in Africa"},
		},
		User{
			"123",
			"Andy Admin",
			"andy_admin@some.org",
			"Admin",
			[]Tag{
				{"Africa", "Locations in Africa"},
			},
		},
	}
}

func TestUser_MakeUpdatedCopy_NoTags(t *testing.T) {
	oldUserData := GetAdminUserData()
	oldObj := oldUserData.User

	newName := "Andrew Admin"
	newEmail := "andrew_admin@some.org"

	newJson := `{
	"Name": "` + newName + `",
	"Email": "` + newEmail + `"
}`
	newObj := oldObj.MakeUpdatedCopy(newJson)

	expected := oldUserData.ID
	results := newObj.ID

	if expected != results {
		t.Errorf("Bad ID.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = newName
	results = newObj.Name
	if expected != results {
		t.Errorf("Bad Name.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = newEmail
	results = newObj.Email
	if expected != results {
		t.Errorf("Bad Email.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = oldUserData.Role
	results = newObj.Role
	if expected != results {
		t.Errorf("Bad Role.  Expected: %s, but got: %s", expected, results)
		return
	}

	expectedTags := oldUserData.Tags
	resultTags := newObj.Tags
	if len(resultTags) != len(expectedTags) {
		t.Errorf("Bad number of Tags.  Expected: %v, but got: %v", len(expectedTags), len(resultTags))
		return
	}

	expectedTag := Tag{"Africa", "Locations in Africa"}
	resultTag := newObj.Tags[0]
	if expectedTag != resultTag {
		t.Errorf("Bad Tag.  Expected: %v, but got: %v", expectedTag, resultTag)
		return
	}
}

func TestUser_MakeUpdatedCopy_EmptyTags(t *testing.T) {
	oldUserData := GetAdminUserData()
	oldObj := oldUserData.User

	newName := "Andrew Admin"
	newEmail := "andrew_admin@some.org"

	newJson := `{
	"Name": "` + newName + `",
	"Email": "` + newEmail + `",
	"Tags" : []
}`
	newObj := oldObj.MakeUpdatedCopy(newJson)

	expected := oldUserData.ID
	results := newObj.ID

	if expected != results {
		t.Errorf("Bad ID.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = newName
	results = newObj.Name
	if expected != results {
		t.Errorf("Bad Name.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = newEmail
	results = newObj.Email
	if expected != results {
		t.Errorf("Bad Email.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = "Admin"
	results = newObj.Role
	if expected != results {
		t.Errorf("Bad Role.  Expected: %s, but got: %s", expected, results)
		return
	}

	expectedTags := []Tag{}

	resultTags := newObj.Tags
	if len(resultTags) != len(expectedTags) {
		t.Errorf("Bad Number of Tags.  Expected: %v, but got: %v", len(expectedTags), len(resultTags))
		return
	}
}

func TestUser_MakeUpdatedCopy_ChangedTags(t *testing.T) {
	oldUserData := GetAdminUserData()
	oldObj := oldUserData.User

	newName := "Andrew Admin"
	newEmail := "andrew_admin@some.org"

	newTag := [2]string{"Asia", "Locations in Asia"}

	newJson := `{
	"Name": "` + newName + `",
	"Email": "` + newEmail + `",
	"Tags": [{"Name": "` + newTag[0] + `", "Description": "` + newTag[1] + `"}]
}`

	newObj := oldObj.MakeUpdatedCopy(newJson)

	expected := oldUserData.ID
	results := newObj.ID

	if expected != results {
		t.Errorf("Bad ID.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = newName
	results = newObj.Name
	if expected != results {
		t.Errorf("Bad Name.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = newEmail
	results = newObj.Email
	if expected != results {
		t.Errorf("Bad Email.  Expected: %s, but got: %s", expected, results)
		return
	}

	expected = "Admin"
	results = newObj.Role
	if expected != results {
		t.Errorf("Bad Role.  Expected: %s, but got: %s", expected, results)
		return
	}

	expectedTags := []Tag{{newTag[0], newTag[1]}}

	resultTags := newObj.Tags
	if len(resultTags) != len(expectedTags) {
		t.Errorf("Bad Number of Tags.  Expected: %v, but got: %v", len(expectedTags), len(resultTags))
		return
	}

	expectedTag := expectedTags[0]
	resultTag := newObj.Tags[0]
	if expectedTag != resultTag {
		t.Errorf("Bad Tag.  Expected: %v, but got: %v", expectedTag, resultTag)
		return
	}
}

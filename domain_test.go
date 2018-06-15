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

func getTestTags() []Tag {
	return []Tag{
		{
			"tag-000",
			"000",
			"",
			"",
		},
		{
			"tag-111",
			"111",
			"",
			"",
		},
		{
			"tag-222",
			"222",
			"",
			"",
		},
		{
			"tag-333",
			"333",
			"",
			"",
		},
	}
}

func TestDoTagsOverlap(t *testing.T) {
	allTags := getTestTags()

	type testData struct {
		tags1    []Tag
		tags2    []Tag
		expected bool
	}

	allTestData := []testData{
		{
			tags1:    []Tag{},
			tags2:    []Tag{},
			expected: false,
		},
		{
			tags1:    []Tag{allTags[0]},
			tags2:    []Tag{},
			expected: false,
		},
		{
			tags1:    []Tag{},
			tags2:    []Tag{allTags[0]},
			expected: false,
		},
		{
			tags1:    []Tag{allTags[0], allTags[1]},
			tags2:    []Tag{allTags[2], allTags[3]},
			expected: false,
		},
		{
			tags1:    []Tag{allTags[0], allTags[1], allTags[2]},
			tags2:    []Tag{allTags[3], allTags[2]},
			expected: true,
		},
	}

	for index, nextData := range allTestData {
		results := DoTagsOverlap(nextData.tags1, nextData.tags2)

		if results != nextData.expected {
			msg := "Bad results for data set %d. Expected %v, but got %v."
			t.Errorf(msg, index, nextData.expected, results)
			break
		}
	}
}

func TestCanUserUseNode(t *testing.T) {
	allTags := getTestTags()
	user := User{
		ID:    "123",
		UID:   "AA123",
		Name:  "Andy Admin",
		Email: "andy_admin@some.org",
		Role:  "admin",
		Tags:  []Tag{allTags[3]},
	}

	node := Node{}
	node.MacAddr = "11:22:33:44:55:66"

	type testData struct {
		nodeTags []Tag
		expected bool
	}

	allTestData := []testData{
		{
			nodeTags: []Tag{allTags[0], allTags[1]},
			expected: false,
		},
		{
			nodeTags: []Tag{allTags[1], allTags[2], allTags[3]},
			expected: true,
		},
	}

	for index, nextData := range allTestData {
		node.Tags = nextData.nodeTags
		results := CanUserUseNode(user, node)

		if results != nextData.expected {
			msg := "Bad results for data set %d. Expected %v, but got %v."
			t.Errorf(msg, index, nextData.expected, results)
			break
		}
	}
}

func TestGetJSONFromSliceEmpty(t *testing.T) {
	testData := []Country{}
	results, err := GetJSONFromSlice(testData)
	expected := "[]"

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if results != expected {
		t.Errorf("Bad results. Expected: %s. But got: %s", expected, results)
	}
}

func TestGetJSONFromSliceGood(t *testing.T) {
	testData := []Country{{Code: "FR", Name: "France"}}
	results, err := GetJSONFromSlice(testData)
	expected := `[{"Code":"FR","Name":"France"}]`

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if results != expected {
		t.Errorf("Bad results. Expected: %s. But got: %s", expected, results)
	}
}

func TestGetJSONFromError(t *testing.T) {
	testData := map[string]string{"A": "aaa"}
	_, err := GetJSONFromSlice(testData)

	if err == nil {
		t.Errorf("Expected an error, but didn't get one")
		return
	}

	expected := "Expected a slice, but got map[A:aaa]."
	results := err.Error()

	if expected != results {
		t.Errorf("Didn't get the expected error message. \nExpected: %s\n But got: %s", expected, results)
	}
}

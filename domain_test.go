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
		tags1    []string
		tags2    []string
		expected bool
	}

	allTestData := []testData{
		{
			[]string{},
			[]string{},
			false,
		},
		{
			[]string{allTags[0].UID},
			[]string{},
			false,
		},
		{
			[]string{},
			[]string{allTags[0].UID},
			false,
		},
		{
			[]string{allTags[0].UID, allTags[1].UID},
			[]string{allTags[2].UID, allTags[3].UID},
			false,
		},
		{
			[]string{allTags[0].UID, allTags[1].UID, allTags[2].UID},
			[]string{allTags[3].UID, allTags[2].UID},
			true,
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
		ID:      "123",
		UID:     "AA123",
		Name:    "Andy Admin",
		Email:   "andy_admin@some.org",
		Role:    "admin",
		TagUIDs: []string{allTags[3].UID},
	}

	node := Node{}
	node.MacAddr = "11:22:33:44:55:66"

	type testData struct {
		nodeTagUIDs []string
		expected    bool
	}

	allTestData := []testData{
		{
			nodeTagUIDs: []string{allTags[0].UID, allTags[1].UID},
			expected:    false,
		},
		{
			nodeTagUIDs: []string{allTags[1].UID, allTags[2].UID, allTags[3].UID},
			expected:    true,
		},
	}

	for index, nextData := range allTestData {
		node.TagUIDs = nextData.nodeTagUIDs
		results := CanUserUseNode(user, node)

		if results != nextData.expected {
			msg := "Bad results for data set %d. Expected %v, but got %v."
			t.Errorf(msg, index, nextData.expected, results)
			break
		}
	}
}

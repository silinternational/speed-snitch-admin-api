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
			"000",
			"Tag 000",
			"",
		},
		{
			"111",
			"Tag 111",
			"",
		},
		{
			"222",
			"Tag 222",
			"",
		},
		{
			"333",
			"Tag 333",
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
		{[]Tag{}, []Tag{}, false},
		{[]Tag{allTags[0]}, []Tag{}, false},
		{[]Tag{}, []Tag{allTags[0]}, false},
		{
			[]Tag{allTags[0], allTags[1]},
			[]Tag{allTags[2], allTags[3]},
			false,
		},
		{
			[]Tag{allTags[0], allTags[1], allTags[2]},
			[]Tag{allTags[3], allTags[2]},
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
		"123",
		"AA123",
		"Andy Admin",
		"andy_admin@some.org",
		"Admin",
		[]Tag{allTags[3]},
	}

	node := Node{}
	node.MacAddr = "11:22:33:44:55:66"

	type testData struct {
		nodeTags []Tag
		expected bool
	}

	allTestData := []testData{
		{
			[]Tag{allTags[0], allTags[1]},
			false,
		},
		{
			[]Tag{allTags[1], allTags[2], allTags[3]},
			true,
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

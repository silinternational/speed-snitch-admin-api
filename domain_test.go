package domain

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
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
			Model: gorm.Model{
				ID: 1,
			},
			Name:        "000",
			Description: "",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Name:        "111",
			Description: "",
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Name:        "222",
			Description: "",
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Name:        "333",
			Description: "",
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

func TestCanUserSeeReportingEvent(t *testing.T) {
	allTags := getTestTags()
	user := User{
		Name:  "Andy Admin",
		Email: "andy_admin@some.org",
		Role:  "admin",
		Tags:  []Tag{allTags[3]},
	}

	node := Node{}
	node.MacAddr = "11:22:33:44:55:66"
	node.ID = 1

	event := ReportingEvent{
		Name:   "Test Event",
		Node:   node,
		NodeID: 1,
	}

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
		event.Node.Tags = nextData.nodeTags
		results := CanUserSeeReportingEvent(user, event)

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

func TestReportingEvent_SetTimestamp(t *testing.T) {

	rEvent := ReportingEvent{
		Date:        "2018-06-26",
		Name:        "Test",
		Description: "Test Reporting Event",
	}

	err := rEvent.SetTimestamp()
	if err != nil {
		t.Errorf("Got unexpected error:\n%s", err.Error())
		return
	}

	expected := int64(1529971200)
	if rEvent.Timestamp != expected {
		t.Errorf("Got wrong timestamp. Expected: %v, but got: %v", expected, rEvent.Timestamp)
	}

	rEvent.Date = "2018-29-29"

	err = rEvent.SetTimestamp()
	if err == nil {
		t.Errorf("Expected an error but didn't get one")
		return
	}
}

func TestGetJSONFromSliceEmpty(t *testing.T) {
	testData := []Country{}
	results, err := GetSliceSafeJSON(testData)
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
	testData := []Country{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Code: "FR",
			Name: "France",
		},
	}
	results, err := GetSliceSafeJSON(testData)
	expected := `[{"ID":1,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"Code":"FR","Name":"France"}]`

	if err != nil {
		t.Errorf("Got an unexpected error: %s", err.Error())
		return
	}

	if results != expected {
		t.Errorf("Bad results. Expected: %s. But got: %s", expected, results)
	}
}

func TestClientError(t *testing.T) {
	body := `abcd`
	results, err := ClientError(1, body)
	if err != nil {
		t.Errorf("Got unexpected error:\n %s", err.Error())
		return
	}

	expected := `{"Error":"abcd"}`
	if results.Body != expected {
		t.Errorf("Bad results. \nExpected: %s. \n But got: %s", expected, results.Body)
		return
	}

	var js map[string]interface{}
	err = json.Unmarshal([]byte(results.Body), &js)

	if err != nil {
		t.Errorf("Results were not valid json. Got error: \n%s", err.Error())
		return
	}
}

func TestCleanBusinessTimes(t *testing.T) {
	// Good - early
	start := "00:00"
	close := "11:59"

	resultStart, resultClose, err := CleanBusinessTimes(start, close)
	if err != nil {
		t.Errorf("Unexpected error.\n%s", err.Error())
		return
	}

	if resultStart != start || resultClose != close {
		t.Errorf("Bad results. Expected: %s and %s, but got %s and %s", start, close, resultStart, resultClose)
		return
	}

	// Good - middle
	start = "08:00"
	close = "14:00"

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err != nil {
		t.Errorf("Unexpected error.\n%s", err.Error())
		return
	}

	if resultStart != start || resultClose != close {
		t.Errorf("Bad results. Expected: %s and %s, but got %s and %s", start, close, resultStart, resultClose)
		return
	}

	// Good - late
	start = "12:00"
	close = "23:59"

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err != nil {
		t.Errorf("Unexpected error.\n%s", err.Error())
		return
	}

	if resultStart != start || resultClose != close {
		t.Errorf("Bad results. Expected: %s and %s, but got %s and %s", start, close, resultStart, resultClose)
		return
	}

	// Bad Formatting
	start = "08-00"
	close = "14-00"

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err == nil {
		t.Error("Expected an error, but didn't get one.")
		return
	}

	// Bad Number
	start = "08:00"
	close = "25:00"

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err == nil {
		t.Error("Expected an error, but didn't get one.")
		return
	}

	// Close time too early
	start = "08:00"
	close = "04:00"

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err == nil {
		t.Error("Expected an error, but didn't get one.")
		return
	}

	// Only start time given
	start = "08:00"
	close = ""

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err == nil {
		t.Error("Expected an error, but didn't get one.")
		return
	}

	// Only close time given
	start = ""
	close = "18:00"

	resultStart, resultClose, err = CleanBusinessTimes(start, close)
	if err == nil {
		t.Error("Expected an error, but didn't get one.")
		return
	}
}

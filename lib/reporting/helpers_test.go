package reporting

import (
	"testing"
	"time"
)

func TestGetStartEndTimestampsForDate(t *testing.T) {
	dateString := "2018-May-6 16:23:13"
	date, err := time.Parse(DateTimeLayout, dateString)
	if err != nil {
		t.Error("Unable to parse date %s", dateString)
		t.Fail()
	}
	startTime, endTime, err := GetStartEndTimestampsForDate(date)
	if startTime != 1525564800 {
		t.Errorf("Did not get expected startTime (%v), got: %v", 1525564800, startTime)
		t.Fail()
	}
	if endTime != 1525651199 {
		t.Errorf("Did not get expected endTime (%v), got: %v", 1525651199, endTime)
		t.Fail()
	}
}

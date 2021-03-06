package reporting

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"testing"
	"time"
)

func TestGetStartEndTimestampsForDate(t *testing.T) {
	dateString := "2018-May-6 16:23:13"
	date, err := time.Parse(DateTimeLayout, dateString)
	if err != nil {
		t.Errorf("Unable to parse date %s", dateString)
		t.Fail()
	}
	startTime, endTime, err := GetStartEndTimestampsForDate(date, "", "")
	if startTime != 1525564800 {
		t.Errorf("Did not get expected startTime (%v), got: %v", 1525564800, startTime)
		t.Fail()
	}
	if endTime != 1525651199 {
		t.Errorf("Did not get expected endTime (%v), got: %v", 1525651199, endTime)
		t.Fail()
	}
}

func TestGetHigherFloat(t *testing.T) {
	fixtures := []struct {
		first  float64
		second float64
		higher float64
	}{
		{
			first:  1.0,
			second: 1.1,
			higher: 1.1,
		},
		{
			first:  0.0001,
			second: 0.00001,
			higher: 0.0001,
		},
		{
			first:  123.456789,
			second: 12.1212,
			higher: 123.456789,
		},
	}

	for _, fix := range fixtures {
		result := GetHigherFloat(fix.first, fix.second)
		if result != fix.higher {
			t.Error("GetHigherFloat did not return expected winner. Got", result, "expected", fix.higher)
			t.Fail()
		}
	}
}

func TestGetLowerFloat(t *testing.T) {
	fixtures := []struct {
		first  float64
		second float64
		lower  float64
	}{
		{
			first:  1.0,
			second: 1.1,
			lower:  1.0,
		},
		{
			first:  0.0001,
			second: 0.00001,
			lower:  0.00001,
		},
		{
			first:  123.456789,
			second: 12.1212,
			lower:  12.1212,
		},
	}

	for _, fix := range fixtures {
		result := GetLowerFloat(fix.first, fix.second)
		if result != fix.lower {
			t.Error("GetLowerFloat did not return expected winner. Got", result, "expected", fix.lower)
			t.Fail()
		}
	}
}

func TestGetLowerLatency(t *testing.T) {
	fixtures := []struct {
		first  float64
		second float64
		lower  float64
	}{
		{
			first:  1.0,
			second: 1.1,
			lower:  1.0,
		},
		{
			first:  0.0001,
			second: 0.00001,
			lower:  0.00001,
		},
		{
			first:  123.456789,
			second: 12.1212,
			lower:  12.1212,
		},
		{
			first:  0.0,
			second: 0.00001,
			lower:  0.00001,
		},
		{
			first:  0.0001,
			second: 0.0,
			lower:  0.0001,
		},
	}

	for _, fix := range fixtures {
		result := GetLowerLatency(fix.first, fix.second)
		if result != fix.lower {
			t.Error("GetLowerLatency did not return expected winner. Got", result, "expected", fix.lower)
			t.Fail()
		}
	}
}

func TestIsValidReportingInterval(t *testing.T) {
	tests := []struct {
		value   string
		success bool
	}{
		{
			value:   domain.ReportingIntervalDaily,
			success: true,
		},
		{
			value:   domain.ReportingIntervalWeekly,
			success: true,
		},
		{
			value:   domain.ReportingIntervalMonthly,
			success: true,
		},
		{
			value:   "day",
			success: false,
		},
	}

	for _, test := range tests {
		result := IsValidReportingInterval(test.value)
		if result != test.success {
			t.Error("Error validating interval for interval value", test.value)
			t.Fail()
		}
	}
}

func TestStringDateToTime(t *testing.T) {
	tests := []struct {
		given    string
		expected int64
	}{
		{
			given:    "2018-08-01",
			expected: 1533081600,
		},
		{
			given:    "2018-08-02",
			expected: 1533168000,
		},
		{
			given:    "2018-08-03",
			expected: 1533254400,
		},
	}

	for _, test := range tests {
		result, err := StringDateToTime(test.given)
		if err != nil {
			t.Error(err)
		}
		if result.Unix() != test.expected {
			t.Errorf("String date was not converted to right time, expected timestamp %v, got %v. Result time obj: %s",
				test.expected, result.Unix(), result.Format(time.RFC3339))
		}
	}
}

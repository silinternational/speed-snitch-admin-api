package reporting

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"time"
)

const DateTimeLayout = "2006-January-2 15:04:05"

func GetYesterday() time.Time {
	today := time.Now().UTC()
	return today.AddDate(0, 0, -1)
}

func GetStartEndTimestampsForDate(date time.Time) (int64, int64, error) {
	startTimeString := fmt.Sprintf("%v-%v-%v 00:00:00", date.Year(), date.Month(), date.Day())
	startTime, err := time.Parse(DateTimeLayout, startTimeString)
	if err != nil {
		return 0, 0, err
	}
	startTimestamp := startTime.Unix()

	endTimeString := fmt.Sprintf("%v-%v-%v 23:59:59", date.Year(), date.Month(), date.Day())
	endTime, err := time.Parse(DateTimeLayout, endTimeString)
	if err != nil {
		return 0, 0, err
	}
	endTimestamp := endTime.Unix()

	return startTimestamp, endTimestamp, nil
}

func GetLowerFloat(first, second float64) float64 {
	if first < second {
		return first
	}
	return second
}

func GetHigherFloat(first, second float64) float64 {
	if first > second {
		return first
	}
	return second
}

func IsValidReportingInterval(needle string) bool {
	haystack := []string{domain.ReportingIntervalDaily, domain.ReportingIntervalWeekly, domain.ReportingIntervalMonthly}
	isValid, _ := domain.InArray(needle, haystack)
	return isValid
}

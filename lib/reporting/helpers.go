package reporting

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"log"
	"time"
)

const DateTimeLayout = "2006-January-2 15:04:05"
const DateLayout = "2006-01-02"

func GetYesterday() time.Time {
	today := time.Now().UTC()
	return today.AddDate(0, 0, -1)
}

func GetStartEndTimestampsForDate(date time.Time, startTimeOfDay, endTimeOfDay string) (int64, int64, error) {
	if startTimeOfDay == "" {
		startTimeOfDay = "00:00:00"
	}
	startTimeString := fmt.Sprintf("%v-%v-%v %s", date.Year(), date.Month(), date.Day(), startTimeOfDay)
	startTime, err := time.Parse(DateTimeLayout, startTimeString)
	if err != nil {
		return 0, 0, err
	}
	startTimestamp := startTime.Unix()

	if endTimeOfDay == "" {
		endTimeOfDay = "23:59:59"
	}
	endTimeString := fmt.Sprintf("%v-%v-%v %s", date.Year(), date.Month(), date.Day(), endTimeOfDay)
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

func GetLowerLatency(first, second float64) float64 {
	if first == 0 {
		return second
	}

	if second == 0 {
		return first
	}

	return GetLowerFloat(first, second)
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

func StringDateToTime(date string) (time.Time, error) {
	timeObj, err := time.Parse(DateLayout, date)
	if err != nil {
		log.Fatal("Error parsing requested config: ", err.Error())
		return time.Time{}, err
	}

	return timeObj, nil
}

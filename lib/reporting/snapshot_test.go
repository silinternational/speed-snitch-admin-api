package reporting

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"strings"
	"testing"
	"time"
)

func areShortPingEntrySlicesEqual(expected, results []domain.ShortPingEntry) string {
	errMsg := fmt.Sprintf("Bad ShortPingEntry results.\nExpected: %v\n But got: %v", expected, results)

	if len(expected) != len(results) {
		return errMsg
	}

	for _, nextExpected := range expected {
		matchFound, _ := domain.InArray(nextExpected, results)
		if !matchFound {
			return errMsg
		}
	}

	return ""
}

func areShortSpeedTestEntrySlicesEqual(expected, results []domain.ShortSpeedTestEntry) string {
	errMsg := fmt.Sprintf("Bad ShortSpeedTestEntry results.\nExpected: %v\n But got: %v", expected, results)

	if len(expected) != len(results) {
		return errMsg
	}

	for _, nextExpected := range expected {
		matchFound, _ := domain.InArray(nextExpected, results)
		if !matchFound {
			return errMsg
		}
	}

	return ""
}

func TestGenerateDailySnapshotsForDate(t *testing.T) {
	db.FlushTables(t)
	fixturesInRange := []domain.TaskLogEntry{
		{
			ID:        "speedTest-aa:aa:aa:aa:aa:aa",
			MacAddr:   "aa:aa:aa:aa:aa:aa",
			Timestamp: 1528145185,
			Upload:    10.0,
			Download:  10.0,
		},
		{
			ID:        "speedTest-aa:aa:aa:aa:aa:aa",
			MacAddr:   "aa:aa:aa:aa:aa:aa",
			Timestamp: 1528145285,
			Upload:    20.0,
			Download:  20.0,
		},
		{
			ID:        "speedTest-aa:aa:aa:aa:aa:aa",
			MacAddr:   "aa:aa:aa:aa:aa:aa",
			Timestamp: 1528145385,
			Upload:    30.0,
			Download:  30.0,
		},
		{
			ID:        "speedTest-aa:aa:aa:aa:aa:aa",
			MacAddr:   "aa:aa:aa:aa:aa:aa",
			Timestamp: 1528145485,
			Upload:    40.0,
			Download:  40.0,
		},
		{
			ID:                "ping-aa:aa:aa:aa:aa:aa",
			MacAddr:           "aa:aa:aa:aa:aa:aa",
			Timestamp:         1528145485,
			Latency:           5,
			PacketLossPercent: 1,
		},
		{
			ID:                "ping-aa:aa:aa:aa:aa:aa",
			MacAddr:           "aa:aa:aa:aa:aa:aa",
			Timestamp:         1528145486,
			Latency:           10,
			PacketLossPercent: 2,
		},
		{
			ID:                "ping-aa:aa:aa:aa:aa:aa",
			MacAddr:           "aa:aa:aa:aa:aa:aa",
			Timestamp:         1528145487,
			Latency:           15,
			PacketLossPercent: 3,
		},
		{
			ID:        "speedTest-11:11:11:11:11:11",
			MacAddr:   "11:11:11:11:11:11",
			Timestamp: 1528145488,
			Upload:    10.0,
			Download:  10.0,
		},
		{
			ID:                "ping-11:11:11:11:11:11",
			MacAddr:           "11:11:11:11:11:11",
			Timestamp:         1528145489,
			Latency:           15,
			PacketLossPercent: 0,
		},
		{
			ID:              "downtime-aa:aa:aa:aa:aa:aa",
			MacAddr:         "aa:aa:aa:aa:aa:aa",
			Timestamp:       1528145490,
			DowntimeSeconds: 240,
		},
		{
			ID:              "downtime-aa:aa:aa:aa:aa:aa",
			MacAddr:         "aa:aa:aa:aa:aa:aa",
			Timestamp:       1528145491,
			DowntimeSeconds: 60,
		},
	}

	// Load fixtures into db
	for _, fix := range fixturesInRange {
		db.PutItem(domain.TaskLogTable, fix)
	}

	// Process daily snapshots
	date, _ := time.Parse(DateTimeLayout, "2018-June-4 20:51:25")
	count, err := GenerateDailySnapshotsForDate(date)
	if err != nil {
		t.Error("Unable to generate daily snapshots:", err)
		return
	}

	if count != 2 {
		t.Error("Wrong number of snapshots created. Should have created 2, got:", count)
		return
	}

	// Get snapshot for MacAddr aa:aa:aa:aa:aa:aa and make sure averages are right
	startTime, endTime, err := GetStartEndTimestampsForDate(date)
	results, err := db.GetSnapshotsForRange(domain.ReportingIntervalDaily, "aa:aa:aa:aa:aa:aa", startTime, endTime)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(results) != 1 {
		t.Error("Not enough results returned, got ", len(results), "expected 1")
	}

	for _, snap := range results {
		if snap.ID == "daily-aa:aa:aa:aa:aa:aa" {
			if snap.UploadAvg != 25.0 {
				t.Errorf("Daily upload average not as expected (25.0), got: %v", snap.UploadAvg)
				t.Fail()
			}
			if snap.UploadMin != 10.0 {
				t.Errorf("Daily upload min not as expected (10.0), got: %v", snap.UploadMin)
				t.Fail()
			}
			if snap.UploadMax != 40.0 {
				t.Errorf("Daily upload max not as expected (40.0), got: %v", snap.UploadMax)
				t.Fail()
			}
			if snap.PacketLossAvg != 2.0 {
				t.Errorf("Daily packet loss avg not as expected (2.0), got: %v", snap.PacketLossAvg)
				t.Fail()
			}
			if snap.PacketLossMin != 1.0 {
				t.Errorf("Daily packet loss min not as expected (1.0), got: %v", snap.PacketLossMin)
				t.Fail()
			}
			if snap.PacketLossMax != 3.0 {
				t.Errorf("Daily packet loss max not as expected (3.0), got: %v", snap.PacketLossMax)
				t.Fail()
			}
			if snap.NetworkDowntimeSeconds != 300 {
				t.Errorf("Daily network downtime seconds not as expected (300), got: %v", snap.NetworkDowntimeSeconds)
				t.Fail()
			}
			if snap.NetworkOutagesCount != 2 {
				t.Errorf("Daily network outages count not as expected (2), got: %v", snap.NetworkOutagesCount)
				t.Fail()
			}

			rawPingResults := snap.RawPingData
			expectedPings := []domain.ShortPingEntry{}
			for _, nextFixture := range fixturesInRange {
				if strings.HasPrefix(nextFixture.ID, "ping-") && strings.HasSuffix(nextFixture.ID, "aa:aa") {
					expectedPings = append(expectedPings, nextFixture.GetShortPingEntry())
				}
			}

			errMsg := areShortPingEntrySlicesEqual(expectedPings, rawPingResults)
			if errMsg != "" {
				t.Error(errMsg)
				break
			}

			rawSpeedTestResults := snap.RawSpeedTestData
			expectedSpeedTests := []domain.ShortSpeedTestEntry{}
			for _, nextFixture := range fixturesInRange {
				if strings.HasPrefix(nextFixture.ID, "speedTest-") && strings.HasSuffix(nextFixture.ID, "aa:aa") {
					expectedSpeedTests = append(expectedSpeedTests, nextFixture.GetShortSpeedTestEntry())
				}
			}

			errMsg = areShortSpeedTestEntrySlicesEqual(expectedSpeedTests, rawSpeedTestResults)
			if errMsg != "" {
				t.Error(errMsg)
				break
			}

			break
		}
	}
}

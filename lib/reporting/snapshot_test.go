package reporting

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"testing"
	"time"
)

func TestGenerateDailySnapshotsForDate(t *testing.T) {
	fixturesInRange := []domain.TaskLogEntry{
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145185,
			Upload:    10.0,
			Download:  10.0,
		},
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145285,
			Upload:    20.0,
			Download:  20.0,
		},
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145385,
			Upload:    30.0,
			Download:  30.0,
		},
		{
			ID:        "speedTest-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145485,
			Upload:    40.0,
			Download:  40.0,
		},
		{
			ID:        "ping-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145485,
			Latency:   5,
		},
		{
			ID:        "ping-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145485,
			Latency:   10,
		},
		{
			ID:        "ping-aa:bb:cc:dd:11:22",
			MacAddr:   "aa:bb:cc:dd:11:22",
			Timestamp: 1528145485,
			Latency:   15,
		},
		{
			ID:        "speedTest-11:22:33:44:55:66",
			MacAddr:   "11:22:33:44:55:66",
			Timestamp: 1528145485,
			Upload:    10.0,
			Download:  10.0,
		},
		{
			ID:        "ping-11:22:33:44:55:66",
			MacAddr:   "11:22:33:44:55:66",
			Timestamp: 1528145485,
			Latency:   15,
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
		t.Fail()
	}

	if count != 2 {
		t.Error("Not enough snapshots created, should have created 2, got:", count)
		t.Fail()
	}

	// Get snapshot for MacAddr aa:bb:cc:dd:11:22 and make sure averages are right
	startTime, endTime, err := GetStartEndTimestampsForDate(date)
	results, err := db.GetDailySnapshotsForRange(startTime, endTime, "")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(results) != 2 {
		t.Error("Not enough results returned, got ", len(results), "expected 2")
	}

	for _, snap := range results {
		if snap.ID == "daily-aa:bb:cc:dd:11:22" {
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
		}
	}
}

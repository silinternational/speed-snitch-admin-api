package reporting

import (
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"testing"
	"time"
)

func TestGenerateDailySnapshotsForDate(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr: "aa:aa:aa:aa:aa:aa",
	}
	db.PutItem(&node1)

	node2 := domain.Node{
		Model: gorm.Model{
			ID: 2,
		},
		MacAddr: "bb:bb:bb:bb:bb:bb",
	}
	db.PutItem(&node2)

	speedInRange := []domain.TaskLogSpeedTest{
		{
			NodeID:    node1.ID,
			Timestamp: 1528145185,
			Upload:    10.0,
			Download:  10.0,
		},
		{
			NodeID:    node1.ID,
			Timestamp: 1528145285,
			Upload:    20.0,
			Download:  20.0,
		},
		{
			NodeID:    node1.ID,
			Timestamp: 1528145385,
			Upload:    30.0,
			Download:  30.0,
		},
		{
			NodeID:    node1.ID,
			Timestamp: 1528145485,
			Upload:    40.0,
			Download:  40.0,
		},
		{
			NodeID:    node2.ID,
			Timestamp: 1528145488,
			Upload:    10.0,
			Download:  10.0,
		},
	}

	for _, i := range speedInRange {
		db.PutItem(&i)
	}

	pingInRange := []domain.TaskLogPingTest{
		{
			NodeID:            node1.ID,
			Timestamp:         1528145485,
			Latency:           5,
			PacketLossPercent: 1,
		},
		{
			NodeID:            node1.ID,
			Timestamp:         1528145486,
			Latency:           10,
			PacketLossPercent: 2,
		},
		{
			NodeID:            node1.ID,
			Timestamp:         1528145487,
			Latency:           15,
			PacketLossPercent: 3,
		},

		{
			NodeID:            node2.ID,
			Timestamp:         1528145489,
			Latency:           15,
			PacketLossPercent: 0,
		},
	}

	for _, i := range pingInRange {
		db.PutItem(&i)
	}

	downtimeInRange := []domain.TaskLogNetworkDowntime{
		{
			NodeID:          node1.ID,
			Timestamp:       1528145490,
			DowntimeSeconds: 240,
		},
		{
			NodeID:          node1.ID,
			Timestamp:       1528145491,
			DowntimeSeconds: 60,
		},
	}

	for _, i := range downtimeInRange {
		db.PutItem(&i)
	}

	restartsInRange := []domain.TaskLogRestart{
		{
			NodeID:    node1.ID,
			Timestamp: 1528145490,
		},
		{
			NodeID:    node1.ID,
			Timestamp: 1528145491,
		},
	}
	for _, i := range restartsInRange {
		db.PutItem(&i)
	}

	// Process daily snapshots
	date, _ := time.Parse(DateTimeLayout, "2018-June-4 20:51:25")
	count, err := GenerateDailySnapshotsForDate(date)
	if err != nil {
		t.Error("Unable to generate daily snapshots:", err)
	}

	if count != 2 {
		t.Error("Not enough snapshots created, should have created 2, got:", count)
	}

	// Get snapshot for MacAddr aa:aa:aa:aa:aa:aa and make sure averages are right
	startTime, endTime, err := GetStartEndTimestampsForDate(date)
	results, err := db.GetSnapshotsForRange(domain.ReportingIntervalDaily, node1.ID, startTime, endTime)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("Not enough results returned, got ", len(results), "expected 1")
	}

	snap := results[0]

	if snap.UploadAvg != 25.0 {
		t.Errorf("Daily upload average not as expected (25.0), got: %v", snap.UploadAvg)
	}
	if snap.UploadMin != 10.0 {
		t.Errorf("Daily upload min not as expected (10.0), got: %v", snap.UploadMin)
	}
	if snap.UploadMax != 40.0 {
		t.Errorf("Daily upload max not as expected (40.0), got: %v", snap.UploadMax)
	}
	if snap.PacketLossAvg != 2.0 {
		t.Errorf("Daily packet loss avg not as expected (2.0), got: %v", snap.PacketLossAvg)
	}
	if snap.PacketLossMin != 1.0 {
		t.Errorf("Daily packet loss min not as expected (1.0), got: %v", snap.PacketLossMin)
	}
	if snap.PacketLossMax != 3.0 {
		t.Errorf("Daily packet loss max not as expected (3.0), got: %v", snap.PacketLossMax)
	}
	if snap.NetworkDowntimeSeconds != 300 {
		t.Errorf("Daily network downtime seconds not as expected (300), got: %v", snap.NetworkDowntimeSeconds)
	}
	if snap.NetworkOutagesCount != 2 {
		t.Errorf("Daily network outages count not as expected (2), got: %v", snap.NetworkOutagesCount)
	}
	if snap.RestartsCount != 2 {
		t.Errorf("Daily restart count not as expected (2), got %v", snap.RestartsCount)
	}

}

package reporting

import (
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/testutils"
	"testing"
	"time"
)

func TestGenerateDailySnapshotsForDateNoBusinessHours(t *testing.T) {
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
	count, err := GenerateDailySnapshotsForDate(date, false)
	if err != nil {
		t.Error("Unable to generate daily snapshots:", err)
	}

	if count != 2 {
		t.Error("Not enough snapshots created, should have created 2, got:", count)
	}

	// Get snapshot for MacAddr aa:aa:aa:aa:aa:aa and make sure averages are right
	startTime, endTime, err := GetStartEndTimestampsForDate(date, "", "")
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
	if snap.SpeedTestDataPoints != 4 {
		t.Errorf(
			"Daily speed test data points count not as expected (4), got %v",
			snap.SpeedTestDataPoints,
		)
	}
	if snap.LatencyDataPoints != 3 {
		t.Errorf(
			"Daily latency data points count not as expected (3), got %v",
			snap.LatencyDataPoints,
		)
	}
	if snap.BizSpeedTestDataPoints != 0 {
		t.Errorf(
			"Daily Business Hours speed test data points count not as expected (0), got %v",
			snap.BizSpeedTestDataPoints,
		)
	}
}

func TestGenerateDailySnapshotsForDateWithBusinessHours(t *testing.T) {
	testutils.ResetDb(t)

	node1 := domain.Node{
		Model: gorm.Model{
			ID: 1,
		},
		MacAddr:           "aa:aa:aa:aa:aa:aa",
		BusinessStartTime: "08:00",
		BusinessCloseTime: "17:00",
	}
	db.PutItem(&node1)

	node2 := domain.Node{
		Model: gorm.Model{
			ID: 2,
		},
		MacAddr: "bb:bb:bb:bb:bb:bb",
	}
	db.PutItem(&node2)

	speedFixtures := []domain.TaskLogSpeedTest{
		{ // Too Early
			NodeID:    node1.ID,
			Timestamp: 1528090000, // 05:26
			Upload:    10.0,
			Download:  10.0,
		},
		{ // In business hours
			NodeID:    node1.ID,
			Timestamp: 1528110000, // 11:00
			Upload:    10.0,
			Download:  40.0,
		},
		{ // In business hours
			NodeID:    node1.ID,
			Timestamp: 1528120000, // 13:46
			Upload:    20.0,
			Download:  50.0,
		},
		{ // In business hours
			NodeID:    node1.ID,
			Timestamp: 1528130000, // 16:33
			Upload:    30.0,
			Download:  60.0,
		},
		{ // Too late
			NodeID:    node1.ID,
			Timestamp: 1528140000, // 19:20
			Upload:    40.0,
			Download:  40.0,
		},
		{ // Other Node
			NodeID:    node2.ID,
			Timestamp: 1528114000,
			Upload:    10.0,
			Download:  10.0,
		},
	}

	for _, i := range speedFixtures {
		db.PutItem(&i)
	}

	pingFixtures := []domain.TaskLogPingTest{
		{ // Too early
			NodeID:            node1.ID,
			Timestamp:         1528000000, // 04:26
			Latency:           5,
			PacketLossPercent: 1,
		},
		{ // In business hours
			NodeID:            node1.ID,
			Timestamp:         1528111000, // 11:16
			Latency:           5,
			PacketLossPercent: 1,
		},
		{ // In business hours
			NodeID:            node1.ID,
			Timestamp:         1528112000, // 11:33
			Latency:           10,
			PacketLossPercent: 2,
		},
		{ // In business hours
			NodeID:            node1.ID,
			Timestamp:         1528113000, //  11:50
			Latency:           15,
			PacketLossPercent: 3,
		},
		{ // Other Node
			NodeID:            node2.ID,
			Timestamp:         1528113000,
			Latency:           15,
			PacketLossPercent: 0,
		},
	}

	for _, i := range pingFixtures {
		db.PutItem(&i)
	}

	downtimeInRange := []domain.TaskLogNetworkDowntime{
		{ // In business hours
			NodeID:          node1.ID,
			Timestamp:       1528112000,
			DowntimeSeconds: 240,
		},
		{ // In business hours
			NodeID:          node1.ID,
			Timestamp:       1528113000,
			DowntimeSeconds: 60,
		},
		{ // Too late
			NodeID:          node1.ID,
			Timestamp:       1528145491,
			DowntimeSeconds: 60,
		},
	}

	for _, i := range downtimeInRange {
		db.PutItem(&i)
	}

	restartsInRange := []domain.TaskLogRestart{
		{ // Too early
			NodeID:    node1.ID,
			Timestamp: 1528000000,
		},
		{ // In business hours
			NodeID:    node1.ID,
			Timestamp: 1528112000,
		},
		{ // In business hours
			NodeID:    node1.ID,
			Timestamp: 1528113000,
		},
	}
	for _, i := range restartsInRange {
		db.PutItem(&i)
	}

	// Process daily snapshots
	date, _ := time.Parse(DateTimeLayout, "2018-June-4 20:51:25")
	count, err := GenerateDailySnapshotsForDate(date, false)
	if err != nil {
		t.Error("Unable to generate daily snapshots:", err)
	}

	if count != 2 {
		t.Error("Not enough snapshots created, should have created 2, got:", count)
	}

	// Get snapshot for MacAddr aa:aa:aa:aa:aa:aa and make sure averages are right
	startTime, endTime, err := GetStartEndTimestampsForDate(date, "", "")
	results, err := db.GetSnapshotsForRange(domain.ReportingIntervalDaily, node1.ID, startTime, endTime)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("Not enough results returned, got ", len(results), "expected 1")
	}

	snap := results[0]

	if snap.BizUploadAvg != 20.0 {
		t.Errorf("Daily biz upload average not as expected (20.0), got: %v", snap.BizUploadAvg)
	}
	if snap.BizUploadMin != 10.0 {
		t.Errorf("Daily biz upload min not as expected (10.0), got: %v", snap.BizUploadMin)
	}
	if snap.BizUploadMax != 30.0 {
		t.Errorf("Daily biz upload max not as expected (30.0), got: %v", snap.BizUploadMax)
		return
	}
	if snap.BizDownloadAvg != 50.0 {
		t.Errorf("Daily biz download average not as expected (20.0), got: %v", snap.BizDownloadAvg)
	}
	if snap.BizDownloadMin != 40.0 {
		t.Errorf("Daily biz download min not as expected (40.0), got: %v", snap.BizDownloadMin)
	}
	if snap.BizDownloadMax != 60.0 {
		t.Errorf("Daily biz dwonload max not as expected (40.0), got: %v", snap.BizDownloadMax)
		return
	}
	if snap.BizPacketLossAvg != 2.0 {
		t.Errorf("Daily biz packet loss avg not as expected (2.0), got: %v", snap.BizPacketLossAvg)
	}
	if snap.BizPacketLossMin != 1.0 {
		t.Errorf("Daily biz packet loss min not as expected (1.0), got: %v", snap.BizPacketLossMin)
	}
	if snap.BizPacketLossMax != 3.0 {
		t.Errorf("Daily biz packet loss max not as expected (3.0), got: %v", snap.BizPacketLossMax)
		return
	}
	if snap.BizNetworkDowntimeSeconds != 300 {
		t.Errorf("Daily biz network downtime seconds not as expected (300), got: %v", snap.BizNetworkDowntimeSeconds)
	}
	if snap.BizNetworkOutagesCount != 2 {
		t.Errorf("Daily biz network outages count not as expected (2), got: %v", snap.BizNetworkOutagesCount)
	}
	if snap.BizRestartsCount != 2 {
		t.Errorf("Daily biz restart count not as expected (2), got %v", snap.BizRestartsCount)
	}
	if snap.BizSpeedTestDataPoints != 3 {
		t.Errorf(
			"Daily biz speed test data points count not as expected (3), got %v",
			snap.BizSpeedTestDataPoints,
		)
	}
	if snap.BizLatencyDataPoints != 3 {
		t.Errorf(
			"Daily biz latency data points count not as expected (3), got %v",
			snap.BizLatencyDataPoints,
		)
	}

}

func TestGenerateDailySnapshotsForThreeDaysForTwoNodes(t *testing.T) {
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

	logs := []domain.TaskLogPingTest{
		{
			Latency:   1,
			NodeID:    node1.ID,
			Timestamp: 1533081600, // 2018-08-01 00:00:00
		},
		{
			Latency:   1,
			NodeID:    node1.ID,
			Timestamp: 1533168000, // 2018-08-02 00:00:00
		},
		{
			Latency:   1,
			NodeID:    node1.ID,
			Timestamp: 1533254400, // 2018-08-03 00:00:00
		},
		{
			Latency:   2,
			NodeID:    node2.ID,
			Timestamp: 1533081600, // 2018-08-01 00:00:00
		},
		{
			Latency:   2,
			NodeID:    node2.ID,
			Timestamp: 1533168000, // 2018-08-02 00:00:00
		},
		{
			Latency:   2,
			NodeID:    node2.ID,
			Timestamp: 1533254400, // 2018-08-03 00:00:00
		},
	}

	for _, i := range logs {
		db.PutItem(&i)
	}

	reportDate, _ := StringDateToTime("2018-08-03")

	snapsCreated, err := GenerateDailySnapshots(reportDate, 3, false)
	if err != nil {
		t.Error(err)
	}

	if snapsCreated != int64(len(logs)) {
		t.Errorf("Not enough daily snapshots created, expected %v, got %v", len(logs), snapsCreated)
	}
}

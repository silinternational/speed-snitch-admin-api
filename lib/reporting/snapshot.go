package reporting

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"os"
	"time"
)

// Iterate through range of days to create snapshots for, and for each date call GenerateDailySnapshotsForDate
// Returns number of snapshots generated and error/nil
func GenerateDailySnapshots(date time.Time, numDaysToProcess int64, forceOverwrite bool) (int64, error) {
	var snapshotsGenerated int64 = 0

	var days int64
	for days = 0; days < numDaysToProcess; days++ {
		dateToBeProcessed := date.AddDate(0, 0, -int(days))
		dateSnapshots, err := GenerateDailySnapshotsForDate(dateToBeProcessed, forceOverwrite)
		snapshotsGenerated += dateSnapshots
		if err != nil {
			return snapshotsGenerated, err
		}
	}

	return snapshotsGenerated, nil
}

// Iterate through all nodes and generate daily snapshots for given date
// Returns number of snapshots created and error/nil
func GenerateDailySnapshotsForDate(date time.Time, forceOverwrite bool) (int64, error) {
	var snapshotsCreated int64 = 0

	var nodes []domain.Node
	err := db.ListItems(&nodes, "id asc")
	if err != nil {
		return 0, err
	}

	for _, n := range nodes {
		created, err := GenerateDailySnapshotForNodeForDate(n, date, forceOverwrite)
		if created {
			snapshotsCreated++
		}
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v - error generating snapshot for node %v - %s. err: %s", date, n.ID, n.Nickname, err.Error())
			return snapshotsCreated, err
		}
	}

	return snapshotsCreated, nil
}

// Generates daily snapshot for the given node/date. If snapshot already exists for node/date it will not be
// regenerated or overwritten unless forceOverwrite is true
// Returns true/false for if a snapshot was created along with an error if one is present
func GenerateDailySnapshotForNodeForDate(node domain.Node, date time.Time, forceOverwrite bool) (bool, error) {
	startTime, endTime, err := GetStartEndTimestampsForDate(date, "", "")
	if err != nil {
		return false, err
	}

	// check for existing snapshot to update, or create new one
	snapshot := domain.ReportingSnapshot{
		Timestamp: startTime,
		NodeID:    node.ID,
		Interval:  domain.ReportingIntervalDaily,
	}
	err = db.FindOne(&snapshot)
	if !gorm.IsRecordNotFoundError(err) && err != nil {
		return false, err
	} else if snapshot.ID != 0 && !forceOverwrite {
		return false, nil
	}

	err = hydrateSnapshotWithPingLogs(&snapshot, node, startTime, endTime)
	if err != nil {
		return false, err
	}

	// Process speed test results
	err = hydrateSnapshotWithSpeedTestLogs(&snapshot, node, startTime, endTime)
	if err != nil {
		return false, err
	}

	err = hydrateSnapshotWithBusinessHourLogs(&snapshot, node, date)
	if err != nil {
		return false, err
	}

	// Track system restart
	var restarts []domain.TaskLogRestart
	err = db.GetTaskLogForRange(&restarts, node.ID, startTime, endTime)
	if err != nil {
		return false, err
	}

	snapshot.RestartsCount = int64(len(restarts))

	// Process network outages
	var outages []domain.TaskLogNetworkDowntime
	err = db.GetTaskLogForRange(&outages, node.ID, startTime, endTime)
	if err != nil {
		return false, err
	}

	for _, i := range outages {
		snapshot.NetworkDowntimeSeconds += i.DowntimeSeconds
	}

	snapshot.NetworkOutagesCount = int64(len(outages))

	err = db.PutItem(&snapshot)
	if err != nil {
		return false, err
	}

	return true, nil
}

func hydrateSnapshotWithPingLogs(
	snapshot *domain.ReportingSnapshot,
	node domain.Node,
	startTime, endTime int64,
) error {

	var pingLogs []domain.TaskLogPingTest
	err := db.GetTaskLogForRange(&pingLogs, node.ID, startTime, endTime)
	if err != nil {
		return err
	}

	if len(pingLogs) == 0 {
		return nil
	}

	snapshot.LatencyTotal = pingLogs[0].Latency
	snapshot.LatencyMax = pingLogs[0].Latency
	snapshot.LatencyMin = pingLogs[0].Latency

	snapshot.PacketLossTotal = pingLogs[0].PacketLossPercent
	snapshot.PacketLossMax = pingLogs[0].PacketLossPercent
	snapshot.PacketLossMin = pingLogs[0].PacketLossPercent

	floatLogCount := float64(len(pingLogs))

	if floatLogCount > 1 {
		for _, pL := range pingLogs[1:] {
			snapshot.LatencyTotal += pL.Latency
			snapshot.LatencyMax = GetHigherFloat(pL.Latency, snapshot.LatencyMax)
			snapshot.LatencyMin = GetLowerLatency(pL.Latency, snapshot.LatencyMin)

			snapshot.PacketLossTotal += pL.PacketLossPercent
			snapshot.PacketLossMax = GetHigherFloat(pL.PacketLossPercent, snapshot.PacketLossMax)
			snapshot.PacketLossMin = GetLowerFloat(pL.PacketLossPercent, snapshot.PacketLossMin)
		}
	}

	snapshot.LatencyAvg = snapshot.LatencyTotal / floatLogCount
	snapshot.PacketLossAvg = snapshot.PacketLossTotal / floatLogCount
	snapshot.LatencyDataPoints = int64(len(pingLogs))

	return nil
}

func hydrateSnapshotWithBusinessHourPingLogs(
	snapshot *domain.ReportingSnapshot,
	node domain.Node,
	businessStartTimestamp, businessCloseTimestamp int64,
) error {
	var pingLogs []domain.TaskLogPingTest
	err := db.GetTaskLogForRange(&pingLogs, node.ID, businessStartTimestamp, businessCloseTimestamp)
	if err != nil {
		return err
	}

	if len(pingLogs) == 0 {
		return nil
	}

	snapshot.BizLatencyTotal = pingLogs[0].Latency
	snapshot.BizLatencyMax = pingLogs[0].Latency
	snapshot.BizLatencyMin = pingLogs[0].Latency

	snapshot.BizPacketLossTotal = pingLogs[0].PacketLossPercent
	snapshot.BizPacketLossMax = pingLogs[0].PacketLossPercent
	snapshot.BizPacketLossMin = pingLogs[0].PacketLossPercent

	floatLogCount := float64(len(pingLogs))

	if floatLogCount > 1 {
		for _, pL := range pingLogs[1:] {
			snapshot.BizLatencyTotal += pL.Latency
			snapshot.BizLatencyMax = GetHigherFloat(pL.Latency, snapshot.BizLatencyMax)
			snapshot.BizLatencyMin = GetLowerLatency(pL.Latency, snapshot.BizLatencyMin)

			snapshot.BizPacketLossTotal += pL.PacketLossPercent
			snapshot.BizPacketLossMax = GetHigherFloat(pL.PacketLossPercent, snapshot.BizPacketLossMax)
			snapshot.BizPacketLossMin = GetLowerLatency(pL.PacketLossPercent, snapshot.BizPacketLossMin)
		}
	}

	snapshot.BizLatencyAvg = snapshot.BizLatencyTotal / floatLogCount
	snapshot.BizPacketLossAvg = snapshot.BizPacketLossTotal / floatLogCount
	snapshot.BizLatencyDataPoints = int64(len(pingLogs))

	return nil
}

func hydrateSnapshotWithSpeedTestLogs(
	snapshot *domain.ReportingSnapshot,
	node domain.Node,
	startTime, endTime int64,
) error {
	var speedLogs []domain.TaskLogSpeedTest
	err := db.GetTaskLogForRange(&speedLogs, node.ID, startTime, endTime)
	if err != nil {
		return err
	}

	if len(speedLogs) == 0 {
		return nil
	}

	snapshot.DownloadTotal = speedLogs[0].Download
	snapshot.DownloadMax = speedLogs[0].Download
	snapshot.DownloadMin = speedLogs[0].Download

	snapshot.UploadTotal += speedLogs[0].Upload
	snapshot.UploadMax = speedLogs[0].Upload
	snapshot.UploadMin = speedLogs[0].Upload

	floatLogCount := float64(len(speedLogs))

	if floatLogCount > 1 {
		for _, i := range speedLogs[1:] {
			snapshot.DownloadTotal += i.Download
			snapshot.DownloadMax = GetHigherFloat(i.Download, snapshot.DownloadMax)
			snapshot.DownloadMin = GetLowerFloat(i.Download, snapshot.DownloadMin)

			snapshot.UploadTotal += i.Upload
			snapshot.UploadMax = GetHigherFloat(i.Upload, snapshot.UploadMax)
			snapshot.UploadMin = GetLowerFloat(i.Upload, snapshot.UploadMin)
		}
	}

	snapshot.DownloadAvg = snapshot.DownloadTotal / floatLogCount
	snapshot.UploadAvg = snapshot.UploadTotal / floatLogCount
	snapshot.SpeedTestDataPoints = int64(len(speedLogs))
	return nil
}

func hydrateSnapshotWithBusinessHourSpeedTestLogs(
	snapshot *domain.ReportingSnapshot,
	node domain.Node,
	businessStartTimestamp, businessCloseTimestamp int64,
) error {

	var speedLogs []domain.TaskLogSpeedTest
	err := db.GetTaskLogForRange(&speedLogs, node.ID, businessStartTimestamp, businessCloseTimestamp)
	if err != nil {
		return err
	}

	if len(speedLogs) == 0 {
		return nil
	}

	snapshot.BizDownloadTotal = speedLogs[0].Download
	snapshot.BizDownloadMax = speedLogs[0].Download
	snapshot.BizDownloadMin = speedLogs[0].Download

	snapshot.BizUploadTotal += speedLogs[0].Upload
	snapshot.BizUploadMax = speedLogs[0].Upload
	snapshot.BizUploadMin = speedLogs[0].Upload

	floatLogCount := float64(len(speedLogs))

	if floatLogCount > 1 {
		for _, i := range speedLogs[1:] {
			snapshot.BizDownloadTotal += i.Download
			snapshot.BizDownloadMax = GetHigherFloat(i.Download, snapshot.BizDownloadMax)
			snapshot.BizDownloadMin = GetLowerFloat(i.Download, snapshot.BizDownloadMin)

			snapshot.BizUploadTotal += i.Upload
			snapshot.BizUploadMax = GetHigherFloat(i.Upload, snapshot.BizUploadMax)
			snapshot.BizUploadMin = GetLowerFloat(i.Upload, snapshot.BizUploadMin)
		}
	}

	snapshot.BizDownloadAvg = snapshot.BizDownloadTotal / floatLogCount
	snapshot.BizUploadAvg = snapshot.BizUploadTotal / floatLogCount
	snapshot.BizSpeedTestDataPoints = int64(len(speedLogs))
	return nil
}

func hydrateSnapshotWithBusinessHourLogs(
	snapshot *domain.ReportingSnapshot,
	node domain.Node,
	date time.Time,
) error {
	start := node.BusinessStartTime
	close := node.BusinessCloseTime

	if (start == "" || start == "00:00") && (close == "" || close == "00:00") {
		return nil
	}

	businessStartTimestamp, businessCloseTimestamp, err := GetStartEndTimestampsForDate(
		date,
		fmt.Sprintf("%s:00", node.BusinessStartTime),
		fmt.Sprintf("%s:00", node.BusinessCloseTime),
	)

	if err != nil {
		return err
	}

	err = hydrateSnapshotWithBusinessHourPingLogs(snapshot, node, businessStartTimestamp, businessCloseTimestamp)
	if err != nil {
		return err
	}

	err = hydrateSnapshotWithBusinessHourSpeedTestLogs(snapshot, node, businessStartTimestamp, businessCloseTimestamp)
	if err != nil {
		return err
	}

	// Track system restart
	var restarts []domain.TaskLogRestart
	err = db.GetTaskLogForRange(&restarts, node.ID, businessStartTimestamp, businessCloseTimestamp)
	if err != nil {
		return err
	}

	snapshot.BizRestartsCount = int64(len(restarts))

	// Process network outages
	var outages []domain.TaskLogNetworkDowntime
	err = db.GetTaskLogForRange(&outages, node.ID, businessStartTimestamp, businessCloseTimestamp)
	if err != nil {
		return err
	}

	for _, i := range outages {
		snapshot.BizNetworkDowntimeSeconds += i.DowntimeSeconds
	}

	snapshot.BizNetworkOutagesCount = int64(len(outages))
	return nil
}

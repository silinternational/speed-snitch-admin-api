package reporting

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"os"
	"time"
)

func GenerateDailySnapshotsForDate(date time.Time) (int64, error) {
	var nodes []domain.Node
	err := db.ListItems(&nodes, "id asc")
	if err != nil {
		return 0, err
	}

	for _, n := range nodes {
		err = GenerateDailySnapshotsForNode(n, date)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v - error generating snapshot for node %v - %s. err: %s", date, n.ID, n.Nickname, err.Error())
		}
	}

	return int64(len(nodes)), nil
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
	if node.BusinessStartTime == "" && node.BusinessCloseTime == "" {
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

func GenerateDailySnapshotsForNode(node domain.Node, date time.Time) error {
	startTime, endTime, err := GetStartEndTimestampsForDate(date, "", "")
	if err != nil {
		return err
	}

	// check for existing snapshot to update, or create new one
	snapshot := domain.ReportingSnapshot{
		Timestamp: startTime,
		NodeID:    node.ID,
		Interval:  domain.ReportingIntervalDaily,
	}
	err = db.FindOne(&snapshot)
	if !gorm.IsRecordNotFoundError(err) && err != nil {
		return err
	}

	err = hydrateSnapshotWithPingLogs(&snapshot, node, startTime, endTime)
	if err != nil {
		return err
	}

	// Process speed test results
	err = hydrateSnapshotWithSpeedTestLogs(&snapshot, node, startTime, endTime)
	if err != nil {
		return err
	}

	err = hydrateSnapshotWithBusinessHourLogs(&snapshot, node, date)
	if err != nil {
		return err
	}

	// Track system restart
	var restarts []domain.TaskLogRestart
	err = db.GetTaskLogForRange(&restarts, node.ID, startTime, endTime)
	if err != nil {
		return err
	}

	snapshot.RestartsCount = int64(len(restarts))

	// Process network outages
	var outages []domain.TaskLogNetworkDowntime
	err = db.GetTaskLogForRange(&outages, node.ID, startTime, endTime)
	if err != nil {
		return err
	}

	for _, i := range outages {
		snapshot.NetworkDowntimeSeconds += i.DowntimeSeconds
	}

	snapshot.NetworkOutagesCount = int64(len(outages))

	return db.PutItem(&snapshot)
}

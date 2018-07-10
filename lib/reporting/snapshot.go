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
		err = GenerateDailySnapshotsForNode(n.Model.ID, date)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v - error generating snapshot for node %v - %s. err: %s", date, n.Model.ID, n.Nickname, err.Error())
		}
	}

	return int64(len(nodes)), nil
}

func GenerateDailySnapshotsForNode(nodeId uint, date time.Time) error {
	startTime, endTime, err := GetStartEndTimestampsForDate(date)
	if err != nil {
		return err
	}

	// check for existing snapshot to update, or create new one
	snapshot := domain.ReportingSnapshot{
		Timestamp: startTime,
		NodeID:    nodeId,
		Interval:  domain.ReportingIntervalDaily,
	}
	err = db.FindOne(&snapshot)
	if !gorm.IsRecordNotFoundError(err) && err != nil {
		return err
	}

	// Process ping/latency tests
	var pingLog []domain.TaskLogPingTest
	err = db.GetTaskLogForRange(&pingLog, nodeId, startTime, endTime)
	if err != nil {
		return err
	}

	// Set high minimum packet loss to be sure we really get the lowest value from results since 0 is valid
	if len(pingLog) > 0 {
		snapshot.LatencyMax = pingLog[0].Latency
		snapshot.LatencyMin = pingLog[0].Latency
		snapshot.PacketLossMax = pingLog[0].PacketLossPercent
		snapshot.PacketLossMin = pingLog[0].PacketLossPercent

		for _, i := range pingLog {
			snapshot.LatencyTotal += i.Latency
			snapshot.LatencyMax = GetHigherFloat(i.Latency, snapshot.LatencyMax)
			snapshot.LatencyMin = GetLowerLatency(i.Latency, snapshot.LatencyMin)

			snapshot.PacketLossTotal += i.PacketLossPercent
			snapshot.PacketLossMax = GetHigherFloat(i.PacketLossPercent, snapshot.PacketLossMax)
			snapshot.PacketLossMin = GetLowerFloat(i.PacketLossPercent, snapshot.PacketLossMin)
		}

		snapshot.LatencyAvg = snapshot.LatencyTotal / float64(len(pingLog))
		snapshot.PacketLossAvg = snapshot.PacketLossTotal / float64(len(pingLog))
	}

	// Process speed test results
	var speedLog []domain.TaskLogSpeedTest
	err = db.GetTaskLogForRange(&speedLog, nodeId, startTime, endTime)
	if err != nil {
		return err
	}

	if len(speedLog) > 0 {
		snapshot.DownloadMax = speedLog[0].Download
		snapshot.DownloadMin = speedLog[0].Download
		snapshot.UploadMax = speedLog[0].Upload
		snapshot.UploadMin = speedLog[0].Upload

		for _, i := range speedLog {
			snapshot.DownloadTotal += i.Download
			snapshot.DownloadMax = GetHigherFloat(i.Download, snapshot.DownloadMax)
			snapshot.DownloadMin = GetLowerFloat(i.Download, snapshot.DownloadMin)

			snapshot.UploadTotal += i.Upload
			snapshot.UploadMax = GetHigherFloat(i.Upload, snapshot.UploadMax)
			snapshot.UploadMin = GetLowerFloat(i.Upload, snapshot.UploadMin)
		}

		snapshot.DownloadAvg = snapshot.DownloadTotal / float64(len(speedLog))
		snapshot.UploadAvg = snapshot.UploadTotal / float64(len(speedLog))
	}

	// Track system restart
	var restarts []domain.TaskLogRestart
	err = db.GetTaskLogForRange(&restarts, nodeId, startTime, endTime)
	if err != nil {
		return err
	}

	snapshot.RestartsCount = int64(len(restarts))

	// Process network outages
	var outages []domain.TaskLogNetworkDowntime
	err = db.GetTaskLogForRange(&outages, nodeId, startTime, endTime)
	if err != nil {
		return err
	}

	for _, i := range outages {
		snapshot.NetworkDowntimeSeconds += i.DowntimeSeconds
	}

	snapshot.NetworkOutagesCount = int64(len(outages))

	return db.PutItem(&snapshot)
}

package reporting

import (
	"fmt"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"sort"
	"time"
)

func GetDailySpeedTestSnapshotForDate(startTime, endTime int64) (map[string]domain.ReportingSnapshot, error) {
	snapshots := map[string]domain.ReportingSnapshot{}

	taskLogs, err := db.GetTaskLogForRange(startTime, endTime, "", []string{domain.TaskTypeSpeedTest})
	if err != nil {
		return snapshots, err
	}

	for _, entry := range taskLogs {
		nodeSnapshot, exists := snapshots[entry.MacAddr]

		if !exists {
			snapshots[entry.MacAddr] = domain.ReportingSnapshot{
				MacAddr:             entry.MacAddr,
				UploadAvg:           entry.Upload,
				UploadMax:           entry.Upload,
				UploadMin:           entry.Upload,
				UploadTotal:         entry.Upload,
				DownloadAvg:         entry.Download,
				DownloadMax:         entry.Download,
				DownloadMin:         entry.Download,
				DownloadTotal:       entry.Download,
				SpeedTestDataPoints: 1,
				RawSpeedTestData:    []domain.ShortSpeedTestEntry{entry.GetShortSpeedTestEntry()},
			}
			continue
		}

		nodeSnapshot.SpeedTestDataPoints++

		// Update update max/min
		nodeSnapshot.UploadMax = GetHigherFloat(entry.Upload, nodeSnapshot.UploadMax)
		nodeSnapshot.UploadMin = GetLowerFloat(entry.Upload, nodeSnapshot.UploadMin)
		nodeSnapshot.DownloadMax = GetHigherFloat(entry.Download, nodeSnapshot.DownloadMax)
		nodeSnapshot.DownloadMin = GetLowerFloat(entry.Download, nodeSnapshot.DownloadMin)

		// Increment totals
		nodeSnapshot.UploadTotal += entry.Upload
		nodeSnapshot.DownloadTotal += entry.Download

		// Calculate average
		nodeSnapshot.UploadAvg = nodeSnapshot.UploadTotal / float64(nodeSnapshot.SpeedTestDataPoints)
		nodeSnapshot.DownloadAvg = nodeSnapshot.DownloadTotal / float64(nodeSnapshot.SpeedTestDataPoints)

		// Keep the raw data
		nodeSnapshot.RawSpeedTestData = append(nodeSnapshot.RawSpeedTestData, entry.GetShortSpeedTestEntry())
		snapshots[entry.MacAddr] = nodeSnapshot
	}

	return snapshots, nil
}

func GetDailyLatencySnapshotForDate(startTime, endTime int64) (map[string]domain.ReportingSnapshot, error) {
	snapshots := map[string]domain.ReportingSnapshot{}

	taskLogs, err := db.GetTaskLogForRange(startTime, endTime, "", []string{domain.TaskTypePing})
	if err != nil {
		return snapshots, err
	}

	for _, entry := range taskLogs {
		nodeSnapshot, exists := snapshots[entry.MacAddr]

		if !exists {
			snapshots[entry.MacAddr] = domain.ReportingSnapshot{
				MacAddr:           entry.MacAddr,
				LatencyAvg:        entry.Latency,
				LatencyMax:        entry.Latency,
				LatencyMin:        entry.Latency,
				LatencyTotal:      entry.Latency,
				LatencyDataPoints: 1,
				PacketLossAvg:     entry.PacketLossPercent,
				PacketLossMax:     entry.PacketLossPercent,
				PacketLossMin:     entry.PacketLossPercent,
				PacketLossTotal:   entry.PacketLossPercent,
				RawPingData:       []domain.ShortPingEntry{entry.GetShortPingEntry()},
			}
			continue
		}

		nodeSnapshot.LatencyDataPoints++

		// Update update max/min
		nodeSnapshot.LatencyMax = GetHigherFloat(entry.Latency, nodeSnapshot.LatencyMax)
		nodeSnapshot.LatencyMin = GetLowerFloat(entry.Latency, nodeSnapshot.LatencyMin)
		nodeSnapshot.PacketLossMax = GetHigherFloat(entry.PacketLossPercent, nodeSnapshot.PacketLossMax)
		nodeSnapshot.PacketLossMin = GetLowerFloat(entry.PacketLossPercent, nodeSnapshot.PacketLossMin)

		// Increment totals
		nodeSnapshot.LatencyTotal += entry.Latency
		nodeSnapshot.PacketLossTotal += entry.PacketLossPercent

		// Calculate average
		nodeSnapshot.LatencyAvg = nodeSnapshot.LatencyTotal / float64(nodeSnapshot.LatencyDataPoints)
		nodeSnapshot.PacketLossAvg = nodeSnapshot.PacketLossTotal / float64(nodeSnapshot.LatencyDataPoints)

		// Keep the raw data
		nodeSnapshot.RawPingData = append(nodeSnapshot.RawPingData, entry.GetShortPingEntry())

		snapshots[entry.MacAddr] = nodeSnapshot
	}

	return snapshots, nil
}

func GetDailyDowntimeSnapshotForDate(startTime, endTime int64) (map[string]domain.ReportingSnapshot, error) {
	snapshots := map[string]domain.ReportingSnapshot{}

	taskLogs, err := db.GetTaskLogForRange(startTime, endTime, "", []string{domain.LogTypeDowntime})
	if err != nil {
		return snapshots, err
	}

	for _, entry := range taskLogs {
		nodeSnapshot, exists := snapshots[entry.MacAddr]

		if !exists {
			snapshots[entry.MacAddr] = domain.ReportingSnapshot{
				MacAddr:                entry.MacAddr,
				NetworkOutagesCount:    1,
				NetworkDowntimeSeconds: entry.DowntimeSeconds,
			}
			continue
		}

		nodeSnapshot.NetworkOutagesCount++
		nodeSnapshot.NetworkDowntimeSeconds += entry.DowntimeSeconds

		snapshots[entry.MacAddr] = nodeSnapshot
	}

	return snapshots, nil
}

func GetDailyRestartSnapshotForDate(startTime, endTime int64) (map[string]domain.ReportingSnapshot, error) {
	snapshots := map[string]domain.ReportingSnapshot{}

	taskLogs, err := db.GetTaskLogForRange(startTime, endTime, "", []string{domain.LogTypeRestart})
	if err != nil {
		return snapshots, err
	}

	for _, entry := range taskLogs {
		nodeSnapshot, exists := snapshots[entry.MacAddr]
		if !exists {
			snapshots[entry.MacAddr] = domain.ReportingSnapshot{
				MacAddr:       entry.MacAddr,
				RestartsCount: 1,
			}
			continue
		}

		nodeSnapshot.RestartsCount++
		snapshots[entry.MacAddr] = nodeSnapshot
	}
	return snapshots, nil
}

func getExistingOrNewSnapshot(
	macAddr string,
	dailySnapshots map[string]domain.ReportingSnapshot,
	startTime, expirationTime int64,
) domain.ReportingSnapshot {
	nodeSnapshot, exists := dailySnapshots[macAddr]

	if exists {
		return nodeSnapshot
	}

	return domain.ReportingSnapshot{
		ID:             domain.ReportingIntervalDaily + "-" + macAddr,
		MacAddr:        macAddr,
		Timestamp:      startTime,
		ExpirationTime: expirationTime,
	}
}

func updateDailySnapshotsWithSpeedtestSnapshots(
	speedtestSnapshots map[string]domain.ReportingSnapshot,
	dailySnapshots map[string]domain.ReportingSnapshot,
	startTime, expirationTime int64,
) {
	for macAddr, speedtestSnap := range speedtestSnapshots {
		nodeSnapshot := getExistingOrNewSnapshot(macAddr, dailySnapshots, startTime, expirationTime)

		nodeSnapshot.UploadAvg = speedtestSnap.UploadAvg
		nodeSnapshot.UploadMax = speedtestSnap.UploadMax
		nodeSnapshot.UploadMin = speedtestSnap.UploadMin
		nodeSnapshot.UploadTotal = speedtestSnap.UploadTotal
		nodeSnapshot.DownloadAvg = speedtestSnap.DownloadAvg
		nodeSnapshot.DownloadMax = speedtestSnap.DownloadMax
		nodeSnapshot.DownloadMin = speedtestSnap.DownloadMin
		nodeSnapshot.DownloadTotal = speedtestSnap.DownloadTotal
		nodeSnapshot.SpeedTestDataPoints = speedtestSnap.SpeedTestDataPoints

		nodeSnapshot.RawSpeedTestData = speedtestSnap.RawSpeedTestData

		dailySnapshots[macAddr] = nodeSnapshot
	}
}

func updateDailySnapshotsWithLatencySnapshots(
	latencySnapshots map[string]domain.ReportingSnapshot,
	dailySnapshots map[string]domain.ReportingSnapshot,
	startTime, expirationTime int64,
) {
	for macAddr, latencySnap := range latencySnapshots {
		nodeSnapshot := getExistingOrNewSnapshot(macAddr, dailySnapshots, startTime, expirationTime)

		nodeSnapshot.LatencyAvg = latencySnap.LatencyAvg
		nodeSnapshot.LatencyMax = latencySnap.LatencyMax
		nodeSnapshot.LatencyMin = latencySnap.LatencyMin
		nodeSnapshot.LatencyTotal = latencySnap.LatencyTotal
		nodeSnapshot.PacketLossMax = latencySnap.PacketLossMax
		nodeSnapshot.PacketLossMin = latencySnap.PacketLossMin
		nodeSnapshot.PacketLossAvg = latencySnap.PacketLossAvg
		nodeSnapshot.PacketLossTotal = latencySnap.PacketLossTotal
		nodeSnapshot.LatencyDataPoints = latencySnap.LatencyDataPoints

		nodeSnapshot.RawPingData = latencySnap.RawPingData

		dailySnapshots[macAddr] = nodeSnapshot
	}
}

func updateDailySnapshotsWithDowntimeSnapshots(
	downtimeSnapshots map[string]domain.ReportingSnapshot,
	dailySnapshots map[string]domain.ReportingSnapshot,
	startTime, expirationTime int64,
) {
	for macAddr, downtimeSnap := range downtimeSnapshots {
		nodeSnapshot := getExistingOrNewSnapshot(macAddr, dailySnapshots, startTime, expirationTime)

		nodeSnapshot.NetworkDowntimeSeconds = downtimeSnap.NetworkDowntimeSeconds
		nodeSnapshot.NetworkOutagesCount = downtimeSnap.NetworkOutagesCount
		dailySnapshots[macAddr] = nodeSnapshot
	}
}

func updateDailySnapshotsWithRestartSnapshots(
	restartSnapshots map[string]domain.ReportingSnapshot,
	dailySnapshots map[string]domain.ReportingSnapshot,
	startTime, expirationTime int64,
) {
	for macAddr, restartSnap := range restartSnapshots {
		nodeSnapshot := getExistingOrNewSnapshot(macAddr, dailySnapshots, startTime, expirationTime)
		nodeSnapshot.RestartsCount = restartSnap.RestartsCount
		dailySnapshots[macAddr] = nodeSnapshot
	}
}

func GenerateDailySnapshotsForDate(date time.Time) (int64, error) {
	startTime, endTime, err := GetStartEndTimestampsForDate(date)
	dailySnapshots := map[string]domain.ReportingSnapshot{}

	speedtestSnapshots, err := GetDailySpeedTestSnapshotForDate(startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("Error getting Daily Speedtest Snapshot. %s", err.Error())
	}

	latencySnapshots, err := GetDailyLatencySnapshotForDate(startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("Error getting Daily Latency Snapshot. %s", err.Error())
	}

	downtimeSnapshots, err := GetDailyDowntimeSnapshotForDate(startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("Error getting Daily Downtime Snapshot. %s", err.Error())
	}

	restartSnapshots, err := GetDailyRestartSnapshotForDate(startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("Error getting Daily Restart Snapshot. %s", err.Error())
	}

	expirationTime := startTime + domain.SecondsPerYear // expire one year after log entry was created

	updateDailySnapshotsWithSpeedtestSnapshots(speedtestSnapshots, dailySnapshots, startTime, expirationTime)
	updateDailySnapshotsWithLatencySnapshots(latencySnapshots, dailySnapshots, startTime, expirationTime)
	updateDailySnapshotsWithDowntimeSnapshots(downtimeSnapshots, dailySnapshots, startTime, expirationTime)
	updateDailySnapshotsWithRestartSnapshots(restartSnapshots, dailySnapshots, startTime, expirationTime)

	// Put snapshots into db
	for _, snapshot := range dailySnapshots {
		// Sort the raw data in place by time stamp before saving to the database
		sort.Slice(snapshot.RawSpeedTestData, func(i, j int) bool {
			return snapshot.RawSpeedTestData[i].Timestamp < snapshot.RawSpeedTestData[j].Timestamp
		})

		sort.Slice(snapshot.RawPingData, func(i, j int) bool {
			return snapshot.RawPingData[i].Timestamp < snapshot.RawPingData[j].Timestamp
		})

		err = db.PutItem(domain.TaskLogTable, snapshot)
		if err != nil {
			return 0, fmt.Errorf("Error saving snapshots to database.\n\t%s\n\t%s", snapshot.MacAddr, err.Error())
		}
	}

	return int64(len(dailySnapshots)), nil
}

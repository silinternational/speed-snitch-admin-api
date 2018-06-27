package reporting

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"sort"
	"strings"
	"time"
)

func GenerateDailySnapshotsForDate(date time.Time) (int64, error) {
	startTime, endTime, err := GetStartEndTimestampsForDate(date)

	taskLogs, err := db.GetTaskLogForRange(startTime, endTime, "", []string{domain.TaskTypePing, domain.TaskTypeSpeedTest})
	if err != nil {
		return 0, err
	}

	dailySnapshots := map[string]domain.ReportingSnapshot{}

	for _, entry := range taskLogs {
		nodeEntry, exists := dailySnapshots[entry.MacAddr]
		if !exists {
			nodeEntry = domain.ReportingSnapshot{
				ID:                  "daily-" + entry.MacAddr,
				Timestamp:           startTime,
				ExpirationTime:      startTime + 31557600, // expire one year after log entry was created
				MacAddr:             entry.MacAddr,
				UploadAvg:           entry.Upload,
				UploadMax:           entry.Upload,
				UploadMin:           entry.Upload,
				UploadTotal:         0.0,
				DownloadAvg:         entry.Download,
				DownloadMax:         entry.Download,
				DownloadMin:         entry.Download,
				DownloadTotal:       0.0,
				LatencyAvg:          entry.Latency,
				LatencyMax:          entry.Latency,
				LatencyMin:          entry.Latency,
				LatencyTotal:        0.0,
				SpeedTestDataPoints: 0,
				LatencyDataPoints:   0,
				RawPingData:         []domain.ShortPingEntry{},
				RawSpeedTestData:    []domain.ShortSpeedTestEntry{},
			}
		}

		if strings.HasPrefix(entry.ID, domain.TaskTypePing) {
			// Increment counts
			nodeEntry.LatencyDataPoints++

			// Update update max/min
			nodeEntry.LatencyMax = GetHigherFloat(entry.Latency, nodeEntry.LatencyMax)
			nodeEntry.LatencyMin = GetLowerFloat(entry.Latency, nodeEntry.LatencyMin)

			// Increment totals
			nodeEntry.LatencyTotal += entry.Latency

			// Calculate average
			nodeEntry.LatencyAvg = nodeEntry.LatencyTotal / float64(nodeEntry.LatencyDataPoints)
			nodeEntry.RawPingData = append(nodeEntry.RawPingData, entry.GetShortPingEntry())

		} else if strings.HasPrefix(entry.ID, domain.TaskTypeSpeedTest) {
			// Increment counts and update max/min
			nodeEntry.SpeedTestDataPoints++

			// Update update max/min
			nodeEntry.UploadMax = GetHigherFloat(entry.Upload, nodeEntry.UploadMax)
			nodeEntry.UploadMin = GetLowerFloat(entry.Upload, nodeEntry.UploadMin)
			nodeEntry.DownloadMax = GetHigherFloat(entry.Download, nodeEntry.DownloadMax)
			nodeEntry.DownloadMin = GetLowerFloat(entry.Download, nodeEntry.DownloadMin)

			// Increment totals
			nodeEntry.UploadTotal += entry.Upload
			nodeEntry.DownloadTotal += entry.Download

			// Calculate average
			nodeEntry.UploadAvg = nodeEntry.UploadTotal / float64(nodeEntry.SpeedTestDataPoints)
			nodeEntry.DownloadAvg = nodeEntry.DownloadTotal / float64(nodeEntry.SpeedTestDataPoints)
			nodeEntry.RawSpeedTestData = append(nodeEntry.RawSpeedTestData, entry.GetShortSpeedTestEntry())

		}

		// Update map
		dailySnapshots[entry.MacAddr] = nodeEntry
	}

	// Put snapshots into db
	for _, snapshot := range dailySnapshots {

		sort.Slice(snapshot.RawSpeedTestData, func(i, j int) bool {
			return snapshot.RawSpeedTestData[i].Timestamp < snapshot.RawSpeedTestData[j].Timestamp
		})

		sort.Slice(snapshot.RawPingData, func(i, j int) bool {
			return snapshot.RawPingData[i].Timestamp < snapshot.RawPingData[j].Timestamp
		})

		err = db.PutItem(domain.TaskLogTable, snapshot)
		if err != nil {
			return 0, err
		}
	}

	return int64(len(dailySnapshots)), nil
}

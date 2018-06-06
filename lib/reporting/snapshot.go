package reporting

import (
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"strings"
	"time"
)

func GenerateDailySnapshotsForDate(date time.Time) (int64, error) {
	startTime, endTime, err := GetStartEndTimestampsForDate(date)

	taskLog, err := db.GetTaskLogForRange(startTime, endTime, "", []string{domain.TaskTypePing, domain.TaskTypeSpeedTest})
	if err != nil {
		return 0, err
	}

	dailySnapshots := map[string]domain.DailySnapshot{}

	for _, entry := range taskLog {
		nodeEntry, exists := dailySnapshots[entry.MacAddr]
		if !exists {
			nodeEntry = domain.DailySnapshot{
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
			}
		}

		// Increment counts and update max/min
		if strings.HasPrefix(entry.ID, domain.TaskTypePing) {
			nodeEntry.LatencyDataPoints++
			if entry.Latency > nodeEntry.LatencyMax {
				nodeEntry.LatencyMax = entry.Latency
			}
			if entry.Latency < nodeEntry.LatencyMin {
				nodeEntry.LatencyMin = entry.Latency
			}
		} else if strings.HasPrefix(entry.ID, domain.TaskTypeSpeedTest) {
			nodeEntry.SpeedTestDataPoints++
			if entry.Upload > nodeEntry.UploadMax {
				nodeEntry.UploadMax = entry.Upload
			}
			if entry.Upload < nodeEntry.UploadMin {
				nodeEntry.UploadMin = entry.Upload
			}

			if entry.Download > nodeEntry.DownloadMax {
				nodeEntry.DownloadMax = entry.Download
			}
			if entry.Download < nodeEntry.DownloadMin {
				nodeEntry.DownloadMin = entry.Download
			}
		}

		// Increment totals
		nodeEntry.UploadTotal += entry.Upload
		nodeEntry.DownloadTotal += entry.Download
		nodeEntry.LatencyTotal += entry.Latency

		// Calculate averages and update min/max
		if nodeEntry.SpeedTestDataPoints > 0 {
			nodeEntry.UploadAvg = nodeEntry.UploadTotal / float64(nodeEntry.SpeedTestDataPoints)
			nodeEntry.DownloadAvg = nodeEntry.DownloadTotal / float64(nodeEntry.SpeedTestDataPoints)
		}
		if nodeEntry.LatencyAvg > 0 {
			nodeEntry.LatencyAvg = nodeEntry.LatencyTotal / float64(nodeEntry.LatencyDataPoints)
		}

		// Update map
		dailySnapshots[entry.MacAddr] = nodeEntry
	}

	// Put snapshots into db
	for _, snapshot := range dailySnapshots {
		err = db.PutItem(domain.TaskLogTable, snapshot)
		if err != nil {
			return 0, err
		}
	}

	return int64(len(dailySnapshots)), nil
}

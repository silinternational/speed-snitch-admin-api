package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/reporting"
	"os"
	"time"
)

type SnapshotConfig struct {
	Date             string `json:"Date"`
	ForceOverwrite   bool   `json:"ForceOverwrite"`
	NumDaysToProcess int64  `json:"NumDaysToProcess"`
}

func handler(config SnapshotConfig) error {
	fmt.Fprintf(os.Stdout, "Starting daily snapshot")

	// Determine what date to start processing snapshots for
	var reportDate time.Time
	if config.Date != "" {
		var err error
		reportDate, err = reporting.StringDateToTime(config.Date)
		if err != nil {
			return err
		}
	} else {
		reportDate = reporting.GetYesterday()
	}

	// Set default number of past days to process to 7 if not already set in config
	if config.NumDaysToProcess == 0 {
		config.NumDaysToProcess = 7
	}

	snapshotCount, err := reporting.GenerateDailySnapshots(reportDate, config.NumDaysToProcess, config.ForceOverwrite)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error generating daily snapshots: %s", err.Error())
		return err
	}

	fmt.Fprintf(os.Stdout, "%v snapshots generated and stored", snapshotCount)

	return nil
}

func main() {
	defer db.Db.Close()
	lambda.Start(handler)
}

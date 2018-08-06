package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"github.com/silinternational/speed-snitch-admin-api/lib/reporting"
	"os"
)

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	fmt.Fprintf(os.Stdout, "Starting daily snapshot")

	yesterday := reporting.GetYesterday()

	snapshotCount, err := reporting.GenerateDailySnapshotsForDate(yesterday)
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

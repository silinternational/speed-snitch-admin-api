package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/lib/speedtestnet"
	"os"
)

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	fmt.Fprintf(os.Stdout, "Starting update speedtestnetservers")
	results, err := speedtestnet.UpdateSTNetServers(domain.SpeedTestNetServerList, speedtestnet.Client{})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Update returned %v stale servers", len(results))

	return nil
}

func main() {
	lambda.Start(handler)
}

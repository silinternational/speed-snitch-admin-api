package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"os"
)

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	fmt.Fprintf(os.Stdout, "Starting database auto migrations")

	err := db.AutoMigrateTables()
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error migrating database: %s", err.Error())
		return err
	}

	return nil
}

func main() {
	defer db.Db.Close()
	lambda.Start(handler)
}

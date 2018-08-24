package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"log"
)

type SnapshotConfig struct {
	DaysMissing int `json:"DaysMissing"`
}

func handler(config SnapshotConfig) []domain.Node {
	log.Println("Starting Alert for MIA Nodes")

	// Set default number of past days to process to 7 if not already set in config
	if config.DaysMissing == 0 {
		config.DaysMissing = 1
	}

	nodes, err := db.ListMIANodes(config.DaysMissing)
	if err != nil {
		log.Println("Error getting list of MIA Nodes: " + err.Error())
		return []domain.Node{}
	}

	log.Printf("%v MIA nodes found", len(nodes))

	return nodes
}

func main() {
	defer db.Db.Close()
	lambda.Start(handler)
}

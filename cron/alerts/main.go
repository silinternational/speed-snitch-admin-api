package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"

	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"log"
	"os"
)

const DaysMissing = int(1)
const SESCharSet = "UTF-8"
const SESSubjectText = "MIA Speedsnitch Nodes"

func getSESReturnToAddr() string {
	envKey := "SES_RETURN_TO_ADDR"
	value := os.Getenv(envKey)
	if value == "" {
		log.Println("Error: required value missing for environment variable " + envKey)
	}
	return value
}

func getSESAWSRegion() string {
	return domain.GetEnv("SES_AWS_REGION", "us-east-1")
}

type AlertsConfig struct {
	DaysMissing     int    `json:"DaysMissing"`
	SESAWSRegion    string `json:"SESAWSRegion"`
	SESCharSet      string `json:"SESCharSet"`
	SESReturnToAddr string `json:"SESReturnToAddr"`
	SESSubjectText  string `json:"SESSubjectText"`
}

func (a *AlertsConfig) setDefaults() {
	if a.DaysMissing == 0 {
		a.DaysMissing = DaysMissing
	}

	if a.SESCharSet == "" {
		a.SESCharSet = SESCharSet
	}

	if a.SESReturnToAddr == "" {
		a.SESReturnToAddr = getSESReturnToAddr()
	}

	if a.SESAWSRegion == "" {
		a.SESAWSRegion = getSESAWSRegion()
	}

	if a.SESSubjectText == "" {
		a.SESSubjectText = SESSubjectText
	}
}

func handler(config AlertsConfig) ([]domain.Node, error) {
	log.Println("Starting Alert for MIA Nodes")

	config.setDefaults()

	nodes, err := db.ListMIANodes(config.DaysMissing)
	if err != nil {
		err := fmt.Errorf("Error getting list of MIA Nodes: %s", err.Error())
		log.Println(err.Error())
		return []domain.Node{}, err
	}

	superAdmins := []domain.User{}
	err = db.ListItems(&superAdmins, "")
	if err != nil {
		err := fmt.Errorf("Error getting list of SuperAdmin users: %s", err.Error())
		log.Println(err.Error())
		return []domain.Node{}, err
	}

	msg := fmt.Sprintf("The following nodes have been MIA for more than %d day(s).", config.DaysMissing)
	scheduledNodes := []domain.Node{}

	for _, node := range nodes {
		if node.IsScheduled() {
			msg = fmt.Sprintf("%s\n%s", node.Nickname)
			scheduledNodes = append(scheduledNodes, node)
		}
	}

	if len(scheduledNodes) < 1 {
		log.Print("No MIA nodes found")
		return scheduledNodes, nil
	}

	charSet := config.SESCharSet

	subject := config.SESSubjectText
	subjContent := ses.Content{
		Charset: &charSet,
		Data:    &subject,
	}

	msgContent := ses.Content{
		Charset: &charSet,
		Data:    &msg,
	}

	msgBody := ses.Body{
		Text: &msgContent,
	}

	recipients := []*string{}
	for _, admin := range superAdmins {
		recipients = append(recipients, aws.String(admin.Email))
	}

	emailMsg := ses.Message{}
	emailMsg.SetSubject(&subjContent)
	emailMsg.SetBody(&msgBody)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: recipients,
		},
		Message: &emailMsg,
		Source:  aws.String(config.SESReturnToAddr),
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.SESAWSRegion)},
	)

	// Create an SES session.
	svc := ses.New(sess)
	result, err := svc.SendEmail(input)
	if err != nil {
		err := fmt.Errorf("Error sending MIA nodes email to superAdmins: %s", err.Error())
		log.Println(err.Error())
		return []domain.Node{}, err
	}

	log.Printf("%v MIA nodes found\n", len(nodes))
	log.Printf("%v MIA node email sent to %v superAdmins\n", len(nodes), len(recipients))
	log.Println(result)

	return scheduledNodes, nil
}

func main() {
	defer db.Db.Close()
	lambda.Start(handler)
}

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
)

const DaysMissing = int(1)
const SESCharSet = "UTF-8"
const SESReturnToAddr = "no_reply@sil.org"
const SESAWSRegion = "us-east-1"
const SESSubjectText = "MIA Speedsnitch Nodes"

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
		a.SESReturnToAddr = SESReturnToAddr
	}

	if a.SESAWSRegion == "" {
		a.SESAWSRegion = SESAWSRegion
	}

	if a.SESSubjectText == "" {
		a.SESSubjectText = SESSubjectText
	}
}

func handler(config AlertsConfig) []domain.Node {
	log.Println("Starting Alert for MIA Nodes")

	config.setDefaults()

	nodes, err := db.ListMIANodes(config.DaysMissing)
	if err != nil {
		log.Println("Error getting list of MIA Nodes: " + err.Error())
		return []domain.Node{}
	}

	superAdmins := []domain.User{}
	err = db.ListItems(&superAdmins, "")
	if err != nil {
		log.Println("Error getting list of SuperAdmin users: " + err.Error())
		return []domain.Node{}
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
		return scheduledNodes
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
		log.Println("Error sending MIA nodes email to superAdmins: " + err.Error())
		return []domain.Node{}
	}

	log.Printf("%v MIA nodes found\n", len(nodes))
	log.Printf("%v MIA node email sent to %v superAdmins\n", len(nodes), len(recipients))
	log.Println(result)

	return scheduledNodes
}

func main() {
	defer db.Db.Close()
	lambda.Start(handler)
}

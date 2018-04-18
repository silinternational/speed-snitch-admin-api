package domain

import (
	"github.com/silinternational/speed-snitch-agent"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"log"
	"os"
)

const TagTable = "tagTable"
const NodeTable = "nodeTable"
const UserTable = "userTable"
const VersionTable = "versionTable"

type Contact struct {
	Name  string `json:"Name"`
	Email string `json:"Email,omitempty"`
	Phone string `json:"Phone,omitempty"`
}

type HelloRequest struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Node struct {
	MacAddr           string       `json:"MacAddr"`
	OS                string       `json:"OS"`
	Arch              string       `json:"Arch"`
	RunningVersion    string       `json:"RunningVersion"`
	ConfiguredVersion string       `json:"ConfiguredVersion"`
	Uptime            string       `json:"Uptime"`
	LastSeen          string       `json:"LastSeen"`
	FirstSeen         string       `json:"FirstSeen"`
	Location          string       `json:"Location"`
	IPAddress         string       `json:"IPAddress"`
	Tasks             []agent.Task `json:"Tasks,omitempty"`
	Contacts          []Contact    `json:"Contacts,omitempty"`
	Tags              []Tag        `json:"Tags,omitempty"`
	ConfiguredBy      string       `json:"ConfiguredBy,omitempty"`
}

type User struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Email string `json:"Email"`
	Role  string `json:"Role"`
	Tags  []Tag  `json:"Tags,omitempty"`
}

type Version struct {
	Number      string `json:"Number"`
	Description string `json:"Description"`
}

var ErrorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

// Add a helper for handling errors. This logs any error to os.Stderr
// and returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	ErrorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func ClientError(status int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       body,
	}, nil
}

// GetTableName returns the env var value of the string passed in
func GetDbTableName(table string) string {
	return os.Getenv(table)
}

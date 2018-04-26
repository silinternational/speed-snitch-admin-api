package domain

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-agent"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const TagTable = "tagTable"
const NodeTable = "nodeTable"
const UserTable = "userTable"
const VersionTable = "versionTable"
const SpeedTestNetServerTable = "speedTestNetServerTable"

const SpeedTestNetServerURL = "http://c.speedtest.net/speedtest-servers-static.php?threads=1"

const UserReqHeaderID = "userID"
const UserRoleSuperAdmin = "superAdmin"

type Contact struct {
	Name  string `json:"Name"`
	Email string `json:"Email,omitempty"`
	Phone string `json:"Phone,omitempty"`
}

type HelloRequest struct {
	ID      string `json:"ID"`
	Version string `json:"Version"`
	Uptime  string `json:"Uptime"`
	OS      string `json:"OS"`
	Arch    string `json:"Arch"`
}

type Tag struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
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
	Coordinates       string       `json:"Coordinates"`
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

type SpeedTestNetServer struct {
	URL         string `xml:"url,attr" json:"URL""`
	Lat         string `xml:"lat,attr" json:"Lat"`
	Lon         string `xml:"lon,attr" json:"Lon"`
	Name        string `xml:"name,attr" json:"Name"`
	Country     string `xml:"country,attr" json:"Country"`
	CountryCode string `xml:"cc,attr"  json:"CountryCode"`
	Sponsor     string `xml:"sponsor,attr" json:"Sponsor"`
	ID          string `xml:"id,attr" json:"ID"`
	URL2        string `xml:"url2,attr" json:"URL2"`
	Host        string `xml:"host,attr" json:"Host"`
}

type STNetServerList struct {
	Servers []SpeedTestNetServer `xml:"speedtestnetserver"`
}

type STNetServerSettings struct {
	ServerLists []STNetServerList `xml:"servers"`
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

// IsValidMacAddress checks whether the input is ...
//   - 12 hexacedimal digits OR
//   - 6 pairs of hexadecimal digits separated by colons and/or hyphens
func IsValidMACAddress(mAddr string) bool {
	controller := "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
	match, _ := regexp.MatchString(controller, mAddr)

	// no separators
	if !match {
		match, _ = regexp.MatchString("^([0-9A-Fa-f]{12})$", mAddr)
	}

	return match
}

func CleanMACAddress(mAddr string) (string, error) {
	if !IsValidMACAddress(mAddr) {
		return "", fmt.Errorf("Invalid MAC Address: " + mAddr)
	}

	return strings.ToLower(mAddr), nil
}

// DoTagsOverlap returns true if there is a tag with the same name
//  in both slices of tags.  Otherwise, returns false.
func DoTagsOverlap(tags1, tags2 []Tag) bool {
	if len(tags1) == 0 || len(tags2) == 0 {
		return false
	}

	for _, tag1 := range tags1 {
		for _, tag2 := range tags2 {
			if tag1.Name == tag2.Name {
				return true
			}
		}
	}

	return false
}

func CanUserUseNode(user User, node Node) bool {
	if user.Role == UserRoleSuperAdmin {
		return true
	}
	return DoTagsOverlap(user.Tags, node.Tags)
}

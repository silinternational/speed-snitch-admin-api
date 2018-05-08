package domain

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-agent"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const DataTable = "dataTable"
const TaskLogTable = "taskLogTable"

const SpeedTestNetServerURL = "http://c.speedtest.net/speedtest-servers-static.php?threads=1"

// Log errors to stderr
var ErrorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

// API call responses have to provide CORS headers manually
var DefaultResponseCorsHeaders = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
}

const UserReqHeaderID = "userID"
const UserRoleSuperAdmin = "superAdmin"
const UserRolerAdmin = "admin"

type Contact struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
	Phone string `json:"Phone"`
}

type HelloRequest struct {
	ID      string `json:"ID"`
	Version string `json:"Version"`
	Uptime  int64  `json:"Uptime"`
	OS      string `json:"OS"`
	Arch    string `json:"Arch"`
}

type Tag struct {
	ID          string `json:"ID"`
	UID         string `json:"UID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

type Node struct {
	ID                string       `json:"ID"`
	MacAddr           string       `json:"MacAddr"`
	OS                string       `json:"OS"`
	Arch              string       `json:"Arch"`
	RunningVersion    string       `json:"RunningVersion"`
	ConfiguredVersion string       `json:"ConfiguredVersion"`
	Uptime            int64        `json:"Uptime"`
	LastSeen          string       `json:"LastSeen"`
	FirstSeen         string       `json:"FirstSeen"`
	Location          string       `json:"Location"`
	Coordinates       string       `json:"Coordinates"`
	Network           string       `json:"Network"`
	IPAddress         string       `json:"IPAddress"`
	Tasks             []agent.Task `json:"Tasks"`
	Contacts          []Contact    `json:"Contacts"`
	TagUIDs           []string     `json:"TagUIDs"`
	ConfiguredBy      string       `json:"ConfiguredBy"`
	Nickname          string       `json:"Nickname"`
	Notes             string       `json:"Notes"`
}

type NodeConfig struct {
	Version struct {
		Number string
		URL    string
	}
	Tasks []agent.Task
}

type User struct {
	ID      string   `json:"ID"`
	UID     string   `json:"UID"`
	UserID  string   `json:"UserID"`
	Name    string   `json:"Name"`
	Email   string   `json:"Email"`
	Role    string   `json:"Role"`
	TagUIDs []string `json:"TagUIDs"`
}

type Version struct {
	ID          string `json:"ID"`
	Number      string `json:"Number"`
	Description string `json:"Description"`
}

type SpeedTestNetServer struct {
	ID          string `json:"ID"`
	URL         string `xml:"url,attr" json:"URL"`
	Lat         string `xml:"lat,attr" json:"Lat"`
	Lon         string `xml:"lon,attr" json:"Lon"`
	Name        string `xml:"name,attr" json:"Name"`
	Country     string `xml:"country,attr" json:"Country"`
	CountryCode string `xml:"cc,attr"  json:"CountryCode"`
	Sponsor     string `xml:"sponsor,attr" json:"Sponsor"`
	ServerID    string `xml:"id,attr" json:"ID"`
	URL2        string `xml:"url2,attr" json:"URL2"`
	Host        string `xml:"host,attr" json:"Host"`
}

type STNetServerList struct {
	Servers []SpeedTestNetServer `xml:"speedtestnetserver"`
}

type STNetServerSettings struct {
	ServerLists []STNetServerList `xml:"servers"`
}

type TaskLogEntry struct {
	ID                 string  `json:"ID"`
	Timestamp          int64   `json:"Timestamp"`
	ExpirationTime     int64   `json:"ExpirationTime"`
	MacAddr            string  `json:"MacAddr"`
	Upload             float64 `json:"Upload"`
	Download           float64 `json:"Download"`
	Latency            float64 `json:"Latency"`
	ErrorCode          string  `json:"ErrorCode"`
	ErrorMessage       string  `json:"ErrorMessage"`
	ServerID           int64   `json:"ServerID"`
	ServerCountry      string  `json:"ServerCountry"`
	ServerCoordinates  string  `json:"ServerCoordinates"`
	ServerSponsor      string  `json:"ServerSponsor"`
	NodeLocation       string  `json:"Location"`
	NodeCoordinates    string  `json:"Coordinates"`
	NodeNetwork        string  `json:"Network"`
	NodeIPAddress      string  `json:"IPAddress"`
	NodeRunningVersion string  `json:"RunningVersion"`
}

// Add a helper for handling errors. This logs any error to os.Stderr
// and returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	ErrorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers:    DefaultResponseCorsHeaders,
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func ClientError(status int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       body,
		Headers:    DefaultResponseCorsHeaders,
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

// GetUrlForAgentVersion creates url to agent binary for given version, os, and arch
func GetUrlForAgentVersion(version, os, arch string) string {
	version = strings.ToLower(version)
	os = strings.ToLower(os)
	arch = strings.ToLower(arch)
	url := fmt.Sprintf(
		"https://github.com/silinternational/speed-snitch-agent/raw/%s/dist/%s/%s/speedsnitch",
		version, os, arch)
	if os == "windows" {
		url = url + ".exe"
	}

	return url
}

// DoTagsOverlap returns true if there is a tag with the same name
//  in both slices of tags.  Otherwise, returns false.
func DoTagsOverlap(tags1, tags2 []string) bool {
	if len(tags1) == 0 || len(tags2) == 0 {
		return false
	}

	for _, tag1 := range tags1 {
		for _, tag2 := range tags2 {
			if tag1 == tag2 {
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
	return DoTagsOverlap(user.TagUIDs, node.TagUIDs)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// GetRandString returns a random string of given length
func GetRandString(length int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// This function will search element inside array with any type.
// Will return boolean and index for matched element.
// True and index more than 0 if element is exist.
// needle is element to search, haystack is slice of value to be search.
func InArray(needle interface{}, haystack interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(haystack)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(needle, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

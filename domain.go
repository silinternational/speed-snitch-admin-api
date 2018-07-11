package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DataTable = "dataTable"
const TaskLogTable = "taskLogTable"

const DataTypeNamedServer = "namedserver"
const DataTypeNode = "node"
const DataTypeSpeedTestNetServer = "speedtestnetserver"
const DataTypeSTNetServerList = "stnetserverlist"
const DataTypeSTNetCountryList = "stnetcountrylist"
const DataTypeTag = "tag"
const DataTypeUser = "user"
const DataTypeVersion = "version"

const LogTypeDowntime = "downtime"
const LogTypeRestart = "restarted"
const LogTypeError = "error"

const ServerTypeSpeedTestNet = "speedTestNet"
const ServerTypeCustom = "custom"

const SpeedTestNetServerList = "http://c.speedtest.net/speedtest-servers-static.php"

const STNetCountryListUID = "1"

const TaskTypePing = "ping"
const TaskTypeSpeedTest = "speedTest"

const TestConfigSpeedTest = "speedTest"
const TestConfigLatencyTest = "latencyTest"

const DefaultPingServerID = "defaultPing"
const DefaultPingServerHost = "paris1.speedtest.orange.fr:8080"

const DefaultSpeedTestNetServerID = "5559"
const DefaultSpeedTestNetServerHost = "paris1.speedtest.orange.fr:8080"

// Log errors to stderr
var ErrorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

const UserReqHeaderUUID = "x-user-uuid"
const UserReqHeaderEmail = "x-user-mail"
const UserRoleSuperAdmin = "superAdmin"
const UserRoleAdmin = "admin"

const PermissionSuperAdmin = "superAdmin"
const PermissionTagBased = "tagBased"

const ReportingIntervalDaily = "daily"
const ReportingIntervalWeekly = "weekly"
const ReportingIntervalMonthly = "monthly"

/***************************************************************
 *
 * Define types that will be stored to database using GORM
 *
 **************************************************************/

type Contact struct {
	gorm.Model
	NodeID uint   `gorm:"not null"`
	Name   string `gorm:"not null"`
	Email  string `gorm:"not null"`
	Phone  string
}

type Country struct {
	gorm.Model
	Code string `gorm:"type:varchar(4);not null;unique_index"`
	Name string `gorm:"type:varchar(64);not null"`
}

type Tag struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `gorm:"not null"`
	Nodes       []Node `gorm:"many2many:node_tags"`
	Users       []User `gorm:"many2many:user_tags"`
}

type Node struct {
	gorm.Model
	MacAddr             string  `gorm:"type:varchar(32);not null;unique_index"`
	OS                  string  `gorm:"type:varchar(16); not null"`
	Arch                string  `gorm:"type:varchar(8); not null"`
	RunningVersion      Version `gorm:"foreignkey:RunningVersionID"`
	RunningVersionID    uint
	ConfiguredVersion   Version `gorm:"foreignkey:ConfiguredVersionID"`
	ConfiguredVersionID uint
	Uptime              int64 `gorm:"default:0"`
	LastSeen            int64 `gorm:"type:int(11)"`
	FirstSeen           int64 `gorm:"type:int(11)"`
	Location            string
	Coordinates         string
	Network             string
	IPAddress           string
	Tasks               []Task
	Contacts            []Contact
	Tags                []Tag `gorm:"many2many:node_tags;"`
	ConfiguredBy        string
	Nickname            string
	Notes               string
}

type Task struct {
	gorm.Model
	NodeID               uint   `gorm:"not null"`
	Type                 string `gorm:"type:varchar(32);not null"`
	Schedule             string `gorm:"not null"`
	NamedServer          NamedServer
	NamedServerID        uint
	SpeedTestNetServerID string
	ServerHost           string
	TaskData             TaskData `gorm:"type:text"`
}

type TaskData struct {
	StringValues map[string]string
	IntValues    map[string]int
	FloatValues  map[string]float64
	IntSlices    map[string][]int
}

func (td TaskData) Value() (driver.Value, error) {
	valueString, err := json.Marshal(td)
	return string(valueString), err
}

func (td *TaskData) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &td)
}

type NamedServer struct {
	gorm.Model
	ServerType           string `gorm:"not null"`
	SpeedTestNetServerID uint   // Only needed if ServerType is SpeedTestNetServer
	SpeedTestNetServer   SpeedTestNetServer
	ServerHost           string // Needed for non-SpeedTestNetServers
	ServerCountry        string
	Name                 string `gorm:"not null"`
	Description          string
	Notes                string `gorm:"type:varchar(2048)"`
}

type User struct {
	gorm.Model
	UUID  string
	Name  string `gorm:"not null"`
	Email string `gorm:"not null;unique_index"`
	Role  string `gorm:"not null"`
	Tags  []Tag  `gorm:"many2many:user_tags"`
}

type Version struct {
	gorm.Model
	Number      string `gorm:"not null;unique_index"`
	Description string `gorm:"not null"`
}

type SpeedTestNetServer struct {
	gorm.Model
	Lat         string `xml:"lat,attr"`
	Lon         string `xml:"lon,attr"`
	Name        string `xml:"name,attr"`
	Country     string `xml:"country,attr"`
	CountryCode string `xml:"cc,attr"`
	ServerID    string `xml:"id,attr" gorm:"not null"`
	Host        string `xml:"host,attr" gorm:"not null"`
}

type TaskLogSpeedTest struct {
	gorm.Model
	Node                 Node
	NodeID               uint    `gorm:"not null"`
	Timestamp            int64   `gorm:"type:int(11); not null"`
	Upload               float64 `gorm:"not null"`
	Download             float64 `gorm:"not null"`
	ServerID             string  `gorm:"not null"`
	ServerCountry        string
	ServerCoordinates    string
	ServerName           string
	NodeLocation         string `gorm:"not null"`
	NodeCoordinates      string `gorm:"not null"`
	NodeNetwork          string
	NodeIPAddress        string  `gorm:"not null"`
	NodeRunningVersion   Version `gorm:"foreignkey:NodeRunningVersionID"`
	NodeRunningVersionID uint    `gorm:"not null"`
}

type TaskLogPingTest struct {
	gorm.Model
	Node                 Node
	NodeID               uint    `gorm:"not null"`
	Timestamp            int64   `gorm:"type:int(11); not null"`
	Latency              float64 `gorm:"not null"`
	PacketLossPercent    float64 `gorm:"not null"`
	ServerID             string
	ServerCountry        string
	ServerCoordinates    string
	ServerName           string
	NodeLocation         string
	NodeCoordinates      string
	NodeNetwork          string
	NodeIPAddress        string
	NodeRunningVersion   Version `gorm:"foreignkey:NodeRunningVersionID"`
	NodeRunningVersionID uint
}

type TaskLogError struct {
	gorm.Model
	Node                 Node
	NodeID               uint
	Timestamp            int64 `gorm:"type:int(11); not null"`
	ErrorCode            string
	ErrorMessage         string
	ServerID             string
	ServerCountry        string
	ServerCoordinates    string
	ServerName           string
	NodeLocation         string
	NodeCoordinates      string
	NodeNetwork          string
	NodeIPAddress        string
	NodeRunningVersion   Version `gorm:"foreignkey:NodeRunningVersionID"`
	NodeRunningVersionID uint
}

type TaskLogRestart struct {
	gorm.Model
	Node      Node
	NodeID    uint
	Timestamp int64 `gorm:"type:int(11); not null"`
}

type TaskLogNetworkDowntime struct {
	gorm.Model
	Node            Node
	NodeID          uint
	Timestamp       int64 `gorm:"type:int(11); not null"`
	DowntimeStart   string
	DowntimeSeconds int64
	NodeNetwork     string
	NodeIPAddress   string
}

type ReportingSnapshot struct {
	gorm.Model
	Node                   Node
	NodeID                 uint  `gorm:"not null"`
	Timestamp              int64 `gorm:"type:int(11); not null"`
	Interval               string
	UploadAvg              float64
	UploadMax              float64
	UploadMin              float64
	UploadTotal            float64
	DownloadAvg            float64
	DownloadMax            float64
	DownloadMin            float64
	DownloadTotal          float64
	LatencyAvg             float64
	LatencyMax             float64
	LatencyMin             float64
	LatencyTotal           float64
	PacketLossAvg          float64
	PacketLossMax          float64
	PacketLossMin          float64
	PacketLossTotal        float64
	SpeedTestDataPoints    int64
	LatencyDataPoints      int64
	NetworkDowntimeSeconds int64
	NetworkOutagesCount    int64
	RestartsCount          int64
}

/***************************************************************
 *
 * Define non-database types
 *
 **************************************************************/

type HelloRequest struct {
	ID      string
	Version string
	Uptime  int64
	OS      string
	Arch    string
}

type NodeConfig struct {
	Version struct {
		Number string
		URL    string
	}
	Tasks []Task
}

type STNetServerList struct {
	Country Country
	Servers []SpeedTestNetServer `xml:"server"`
}

// This relates to the xml response from the external url where we get the list of speedtest.net servers
type STNetServerSettings struct {
	ServerLists []STNetServerList `xml:"servers"`
}

// Add a helper for handling errors. This logs any error to os.Stderr
// and returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	ErrorLogger.Println(err.Error())
	js, _ := json.Marshal(http.StatusText(http.StatusInternalServerError))
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       string(js),
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func ClientError(status int, body string) (events.APIGatewayProxyResponse, error) {
	js, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       string(js),
	}, nil
}

// GetTableName returns the env var value of the string passed in or the string itself
func GetDbTableName(table string) string {
	envOverride := os.Getenv(table)
	if envOverride != "" {
		return envOverride
	}

	return table
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
func GetUrlForAgentVersion(version, operatingsystem, arch string) string {
	downloadBaseUrl := os.Getenv("downloadBaseUrl")
	version = strings.ToLower(version)
	operatingsystem = strings.ToLower(operatingsystem)
	arch = strings.ToLower(arch)
	url := fmt.Sprintf(
		"%s/%s/%s/%s/speedsnitch",
		downloadBaseUrl, version, operatingsystem, arch)
	if operatingsystem == "windows" {
		url = url + ".exe"
	}

	return url
}

// DoTagsOverlap returns true if there is a tag with the same UID
//  in both slices of tags.  Otherwise, returns false.
func DoTagsOverlap(tags1, tags2 []Tag) bool {
	if len(tags1) == 0 || len(tags2) == 0 {
		return false
	}

	for _, tag1 := range tags1 {
		for _, tag2 := range tags2 {
			if tag1.Model.ID == tag2.Model.ID {
				return true
			}
		}
	}

	return false
}

// CanUserUseNode returns true if the user has a superAdmin role or
//   if the user has a tag that the node has
func CanUserUseNode(user User, node Node) bool {
	if user.Role == UserRoleSuperAdmin {
		return true
	}
	return DoTagsOverlap(user.Tags, node.Tags)
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

// GetSliceSafeJSON handles special logic for slices. If the length is 0, returns "[]".
// Otherwise, returns the results of json.Marshal(s)
func GetSliceSafeJSON(v interface{}) (string, error) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(v)

		if s.Len() == 0 {
			return "[]", nil
		}
		js, err := json.Marshal(v)
		if err != nil {
			return "", err
		}

		return string(js), nil
	}

	js, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

func GetUintFromString(param string) uint {
	id, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		id = 0
	}
	return uint(id)
}

func ReturnJsonOrError(response interface{}, err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusNotFound,
				Body:       "",
			}, nil
		}
		return ServerError(err)
	}

	js, err := GetSliceSafeJSON(response)
	if err != nil {
		return ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       js,
	}, nil
}

func GetEnv(name, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

// Get ID from path paramters as uint, otherwise return 0
func GetResourceIDFromRequest(req events.APIGatewayProxyRequest) uint {
	if req.PathParameters["id"] == "" {
		return 0
	}

	id := GetUintFromString(req.PathParameters["id"])
	return id
}

package domain

import (
	"bytes"
	"database/sql/driver"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DataTypeSpeedTestNetServer = "speedtestnetserver"
const DataTypeSTNetServerList = "stnetserverlist"

const LogTypeDowntime = "downtime"
const LogTypeRestart = "restarted"
const LogTypeError = "error"

const ServerTypeSpeedTestNet = "speedTest"
const ServerTypePing = "ping"

const SpeedTestNetServerList = "http://c.speedtest.net/speedtest-servers-static.php"

const TaskTypePing = "ping"
const TaskTypeSpeedTest = "speedTest"

const TestConfigSpeedTest = "speedTest"

const SecondsPerDay = 86400 // 60 * 60 * 24
const BusinessTimeFormat = "15:04"

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

const DateLayout = "2006-01-02"

/***************************************************************
/*
/* Define types that will be stored to database using GORM
/*
/**************************************************************/
type Contact struct {
	gorm.Model
	NodeID uint
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
	Name        string `gorm:"not null;unique_index"`
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
	RunningVersionID    uint    `gorm:"default:null"`
	ConfiguredVersion   Version `gorm:"foreignkey:ConfiguredVersionID" json:"-"`
	ConfiguredVersionID uint    `gorm:"default:null"`
	Uptime              int64   `gorm:"default:0"`
	LastSeen            string  `gorm:"type:varchar(64)"`
	FirstSeen           string  `gorm:"type:varchar(64)"`
	Location            string
	Coordinates         string
	Network             string
	IPAddress           string
	Tasks               []Task
	Contacts            []Contact
	Tags                []Tag `gorm:"many2many:node_tags"`
	ConfiguredBy        string
	Nickname            string
	Notes               string
	BusinessStartTime   string `gorm:"type:varchar(5)"`
	BusinessCloseTime   string `gorm:"type:varchar(5)"`
}

func (n *Node) IsScheduled() bool {
	if len(n.Tasks) > 0 {
		return true
	}
	return false
}

type NodeTags struct {
	gorm.Model
	Tag    Node `gorm:"foreignkey:TagID"`
	TagID  uint
	Node   Node `gorm:"foreignkey:NodeID"`
	NodeID uint
}

type Task struct {
	gorm.Model
	NodeID        uint
	Type          string `gorm:"type:varchar(32);not null"`
	Schedule      string `gorm:"not null"`
	NamedServer   NamedServer
	NamedServerID uint `gorm:"default:null"`
	ServerHost    string
	TaskData      TaskData `gorm:"type:text"`
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
	ServerType           string             `gorm:"not null" json:"Type"`
	SpeedTestNetServerID uint               `gorm:"default:null"` // Only needed if ServerType is SpeedTestNetServer
	SpeedTestNetServer   SpeedTestNetServer `json:"-"`
	ServerHost           string             `json:"Host"` // Needed for non-SpeedTestNetServers
	ServerCountry        string             `json:"Country"`
	ServerCountryCode    string             `gorm:"-" json:"CountryCode"` // Only needed if ServerType is SpeedTestNetServer
	Name                 string             `gorm:"not null;unique_index"`
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

type UserTags struct {
	gorm.Model
	Tag    Node `gorm:"foreignkey:TagID"`
	TagID  uint
	User   User `gorm:"foreignkey:UserID"`
	UserID uint
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
	NamedServer          NamedServer
	NodeID               uint    `gorm:"default:null"`
	Timestamp            int64   `gorm:"type:int(11); not null"`
	Upload               float64 `gorm:"not null;default:0"`
	Download             float64 `gorm:"not null;default:0"`
	NamedServerID        uint    `gorm:"default:null"`
	ServerHost           string
	ServerCountry        string
	ServerCoordinates    string
	ServerName           string
	NodeLocation         string `gorm:"not null"`
	NodeCoordinates      string `gorm:"not null"`
	NodeNetwork          string
	NodeIPAddress        string  `gorm:"not null"`
	NodeRunningVersion   Version `gorm:"foreignkey:NodeRunningVersionID"`
	NodeRunningVersionID uint    `gorm:"default:null"`
}

func (t TaskLogSpeedTest) GetTaskLogMap() map[string]string {
	taskLogMap := map[string]string{
		"NodeID":             fmt.Sprintf("%v", t.NodeID),
		"Date and Time":      TimestampToHumanReadable(t.Timestamp),
		"Upload":             fmt.Sprintf("%.3f", t.Upload),
		"Download":           fmt.Sprintf("%.3f", t.Download),
		"NamedServerID":      fmt.Sprintf("%v", t.NamedServerID),
		"ServerHost":         t.ServerHost,
		"ServerCountry":      t.ServerCountry,
		"ServerName":         t.ServerName,
		"NodeLocation":       t.NodeLocation,
		"NodeCoordinates":    t.NodeCoordinates,
		"NodeNetwork":        t.NodeNetwork,
		"NodeIPAddress":      t.NodeIPAddress,
		"NodeRunningVersion": t.NodeRunningVersion.Number,
	}

	return taskLogMap
}

func (t TaskLogSpeedTest) GetTaskLogKeys() []string {
	taskLogKeys := []string{
		"NodeID",
		"Date and Time",
		"Upload",
		"Download",
	}
	taskLogKeys = append(taskLogKeys, getSharedTaskLogKeys()...)

	return taskLogKeys
}

type TaskLogPingTest struct {
	gorm.Model
	Node                 Node
	NamedServer          NamedServer
	NodeID               uint    `gorm:"default:null"`
	Timestamp            int64   `gorm:"type:int(11); not null"`
	Latency              float64 `gorm:"not null;default:0"`
	PacketLossPercent    float64 `gorm:"not null;default:0"`
	NamedServerID        uint    `gorm:"default:null"`
	ServerHost           string
	ServerCountry        string
	ServerName           string
	NodeLocation         string
	NodeCoordinates      string
	NodeNetwork          string
	NodeIPAddress        string
	NodeRunningVersion   Version `gorm:"foreignkey:NodeRunningVersionID"`
	NodeRunningVersionID uint    `gorm:"default:null"`
}

func (t TaskLogPingTest) GetTaskLogMap() map[string]string {
	taskLogMap := map[string]string{
		"NodeID":             fmt.Sprintf("%v", t.NodeID),
		"Date and Time":      TimestampToHumanReadable(t.Timestamp),
		"Latency":            fmt.Sprintf("%.3f", t.Latency),
		"PacketLossPercent":  fmt.Sprintf("%.3f", t.PacketLossPercent),
		"NamedServerID":      fmt.Sprintf("%v", t.NamedServerID),
		"ServerHost":         t.ServerHost,
		"ServerCountry":      t.ServerCountry,
		"ServerName":         t.ServerName,
		"NodeLocation":       t.NodeLocation,
		"NodeCoordinates":    t.NodeCoordinates,
		"NodeNetwork":        t.NodeNetwork,
		"NodeIPAddress":      t.NodeIPAddress,
		"NodeRunningVersion": t.NodeRunningVersion.Number,
	}

	return taskLogMap
}

func (t TaskLogPingTest) GetTaskLogKeys() []string {
	taskLogKeys := []string{
		"NodeID",
		"Date and Time",
		"Latency",
		"PacketLossPercent",
	}
	taskLogKeys = append(taskLogKeys, getSharedTaskLogKeys()...)
	return taskLogKeys
}

type TaskLogError struct {
	gorm.Model
	Node                 Node
	NodeID               uint  `gorm:"default:null"`
	Timestamp            int64 `gorm:"type:int(11); not null"`
	ErrorCode            string
	ErrorMessage         string
	NamedServer          NamedServer
	NamedServerID        uint `gorm:"default:null"`
	ServerHost           string
	ServerCountry        string
	ServerName           string
	NodeLocation         string
	NodeCoordinates      string
	NodeNetwork          string
	NodeIPAddress        string
	NodeRunningVersion   Version `gorm:"foreignkey:NodeRunningVersionID"`
	NodeRunningVersionID uint    `gorm:"default:null"`
}

type TaskLogRestart struct {
	gorm.Model
	Node      Node
	NodeID    uint  `gorm:"default:null"`
	Timestamp int64 `gorm:"type:int(11); not null"`
}

func (t TaskLogRestart) GetTaskLogMap() map[string]string {
	taskLogMap := map[string]string{
		"NodeID":        fmt.Sprintf("%v", t.NodeID),
		"Date and Time": TimestampToHumanReadable(t.Timestamp),
	}

	return taskLogMap
}

func (t TaskLogRestart) GetTaskLogKeys() []string {
	taskLogKeys := []string{
		"NodeID",
		"Date and Time",
	}
	return taskLogKeys
}

type TaskLogNetworkDowntime struct {
	gorm.Model
	Node            Node
	NodeID          uint   `gorm:"default:null"`
	Timestamp       int64  `gorm:"type:int(11); not null;default:0"`
	DowntimeStart   string `gorm:"not null"`
	DowntimeSeconds int64  `gorm:"not null;default:0"`
	NodeNetwork     string
	NodeIPAddress   string
}

func (t TaskLogNetworkDowntime) GetTaskLogMap() map[string]string {
	taskLogMap := map[string]string{
		"NodeID":          fmt.Sprintf("%v", t.NodeID),
		"Date and Time":   TimestampToHumanReadable(t.Timestamp),
		"DowntimeStart":   t.DowntimeStart,
		"DowntimeSeconds": fmt.Sprintf("%v", t.DowntimeSeconds),
		"NodeNetwork":     t.NodeNetwork,
		"NodeIPAddress":   t.NodeIPAddress,
	}

	return taskLogMap
}

func (t TaskLogNetworkDowntime) GetTaskLogKeys() []string {
	taskLogKeys := []string{
		"NodeID",
		"Date and Time",
		"DowntimeStart",
		"DowntimeSeconds",
		"NodeNetwork",
		"NodeIPAddress",
	}
	return taskLogKeys
}

type ReportingSnapshot struct {
	gorm.Model
	Node                      Node
	NodeID                    uint    `gorm:"default:null"`
	Timestamp                 int64   `gorm:"type:int(11); not null"`
	Interval                  string  `gorm:"not null"`
	UploadAvg                 float64 `gorm:"not null;default:0"`
	UploadMax                 float64 `gorm:"not null;default:0"`
	UploadMin                 float64 `gorm:"not null;default:0"`
	UploadTotal               float64 `gorm:"not null;default:0"`
	DownloadAvg               float64 `gorm:"not null;default:0"`
	DownloadMax               float64 `gorm:"not null;default:0"`
	DownloadMin               float64 `gorm:"not null;default:0"`
	DownloadTotal             float64 `gorm:"not null;default:0"`
	LatencyAvg                float64 `gorm:"not null;default:0"`
	LatencyMax                float64 `gorm:"not null;default:0"`
	LatencyMin                float64 `gorm:"not null;default:0"`
	LatencyTotal              float64 `gorm:"not null;default:0"`
	PacketLossAvg             float64 `gorm:"not null;default:0"`
	PacketLossMax             float64 `gorm:"not null;default:0"`
	PacketLossMin             float64 `gorm:"not null;default:0"`
	PacketLossTotal           float64 `gorm:"not null;default:0"`
	SpeedTestDataPoints       int64   `gorm:"not null;default:0"`
	LatencyDataPoints         int64   `gorm:"not null;default:0"`
	NetworkDowntimeSeconds    int64   `gorm:"not null;default:0"`
	NetworkOutagesCount       int64   `gorm:"not null;default:0"`
	RestartsCount             int64   `gorm:"not null;default:0"`
	BizUploadAvg              float64 `gorm:"not null;default:0"`
	BizUploadMax              float64 `gorm:"not null;default:0"`
	BizUploadMin              float64 `gorm:"not null;default:0"`
	BizUploadTotal            float64 `gorm:"not null;default:0"`
	BizDownloadAvg            float64 `gorm:"not null;default:0"`
	BizDownloadMax            float64 `gorm:"not null;default:0"`
	BizDownloadMin            float64 `gorm:"not null;default:0"`
	BizDownloadTotal          float64 `gorm:"not null;default:0"`
	BizLatencyAvg             float64 `gorm:"not null;default:0"`
	BizLatencyMax             float64 `gorm:"not null;default:0"`
	BizLatencyMin             float64 `gorm:"not null;default:0"`
	BizLatencyTotal           float64 `gorm:"not null;default:0"`
	BizPacketLossAvg          float64 `gorm:"not null;default:0"`
	BizPacketLossMax          float64 `gorm:"not null;default:0"`
	BizPacketLossMin          float64 `gorm:"not null;default:0"`
	BizPacketLossTotal        float64 `gorm:"not null;default:0"`
	BizSpeedTestDataPoints    int64   `gorm:"not null;default:0"`
	BizLatencyDataPoints      int64   `gorm:"not null;default:0"`
	BizNetworkDowntimeSeconds int64   `gorm:"not null;default:0"`
	BizNetworkOutagesCount    int64   `gorm:"not null;default:0"`
	BizRestartsCount          int64   `gorm:"not null;default:0"`
}

type ReportingEvent struct {
	gorm.Model
	Node        Node   `gorm:"foreignkey:NodeID" json:"-"`
	NodeID      uint   `gorm:"default:null;unique_index:idx_node_name_date"`
	Timestamp   int64  `gorm:"type:int(11); not null;default:0"`
	Date        string `gorm:"not null;unique_index:idx_node_name_date"`
	Name        string `gorm:"not null;unique_index:idx_node_name_date"`
	Description string `gorm:"type:varchar(2048)"`
}

func (r *ReportingEvent) SetTimestamp() error {

	timestamp, err := time.Parse(DateLayout, r.Date)

	if err != nil {
		errMsg := fmt.Sprintf("Error interpreting date %s. %s", r.Date, err.Error())
		return fmt.Errorf(errMsg)
	}

	r.Timestamp = timestamp.Unix()
	return nil
}

/***************************************************************
/*
/* Define non-database types
/*
/**************************************************************/

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

type AssociationReplacements struct {
	Replacements    interface{}
	AssociationName string
}

type STNetServerList struct {
	Country Country
	Servers []SpeedTestNetServer `xml:"server"`
}

// This relates to the xml response from the external url where we get the list of speedtest.net servers
type STNetServerSettings struct {
	ServerLists []STNetServerList `xml:"servers"`
}

type ForeignKey struct {
	ChildModel  interface{}
	ChildField  string
	ParentTable string
	ParentField string
	OnDelete    string
	OnUpdate    string
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
	}, err
}

// Similarly add a helper for send responses relating to client errors.
func ClientError(status int, body string) (events.APIGatewayProxyResponse, error) {

	type cError struct {
		Error string
	}

	js, _ := json.Marshal(cError{Error: body})

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       string(js),
	}, nil
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

// CleanBusinessTimes takes to strings for 24-hour times (HH:MM).
// If they are both empty strings, it returns empty strings.
// Otherwise makes sure there is no error parsing them into time.Time values,
// in which case it returns the original values
func CleanBusinessTimes(start, close string) (string, string, error) {
	if start == "" && close == "" {
		return start, close, nil
	}

	// If only one is set, error
	if start == "" || close == "" {
		errMsg := fmt.Sprintf(
			`Error with business hours.  If one value is set, the other must also be set.\n Got "%s" and "%s".`,
			start, close)
		return start, close, fmt.Errorf("Error parsing business start time.\n %s", errMsg)
	}

	startTime, err := time.Parse(BusinessTimeFormat, start)
	if err != nil {
		return start, close, fmt.Errorf("Error parsing business start time.\n %s", err.Error())
	}

	closeTime, err := time.Parse(BusinessTimeFormat, close)
	if err != nil {
		return start, close, fmt.Errorf("Error parsing business close time.\n %s", err.Error())
	}

	if !closeTime.After(startTime) {
		return start, close, fmt.Errorf(
			"Error parsing business times. A 24-hour format must be used and close time must come after start time.",
		)
	}

	return start, close, nil
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
			if tag1.ID == tag2.ID {
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

// CanUserSeeReportingEvent returns true if the user has a superAdmin role or
//   if the event has no node associated with it or
//   if the user has a tag that the event's node has
func CanUserSeeReportingEvent(user User, event ReportingEvent) bool {
	if user.Role == UserRoleSuperAdmin || event.NodeID == 0 {
		return true
	}

	return DoTagsOverlap(user.Tags, event.Node.Tags)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

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

type TaskLogMapper interface {
	GetTaskLogMap() map[string]string
	GetTaskLogKeys() []string
}

func ReturnCSVOrError(items []TaskLogMapper, filename string, err error) (events.APIGatewayProxyResponse, error) {
	if len(items) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "[]",
		}, nil
	}

	var b bytes.Buffer
	csvWriter := csv.NewWriter(&b)

	defer csvWriter.Flush()

	columnKeys := items[0].GetTaskLogKeys()

	// Write the Column Headers
	err = csvWriter.Write(columnKeys)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "",
		}, err
	}

	for _, item := range items {
		itemMap := item.GetTaskLogMap()
		nextRow := []string{}
		for _, key := range columnKeys {
			nextRow = append(nextRow, itemMap[key])
		}

		err := csvWriter.Write(nextRow)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "",
			}, err
		}
	}

	csvWriter.Flush()
	csvOutput := b.String()

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       csvOutput,
		Headers: map[string]string{
			"Content-Type":        "text/csv",
			"Content-Disposition": "attachment;filename=" + filename,
		},
	}

	return response, nil
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

// GetTimeNow returns the current UTC time in the RFC3339 format
func GetTimeNow() string {
	t := time.Now().UTC()
	return t.Format(time.RFC3339)
}

func TimestampToHumanReadable(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(time.RFC3339)
}

func getSharedTaskLogKeys() []string {
	keys := []string{
		"NamedServerID",
		"ServerHost",
		"ServerCountry",
		"ServerName",
		"NodeLocation",
		"NodeCoordinates",
		"NodeNetwork",
		"NodeIPAddress",
		"NodeRunningVersion",
	}
	return keys
}

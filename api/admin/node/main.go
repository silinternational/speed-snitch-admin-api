package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

const SelfType = domain.DataTypeNode

const DefaultPingTimeoutInSeconds = 5
const DefaultSpeedTestTimeoutInSeconds = 60 // 1 minute
const DefaultSpeedTestMaxSeconds = 60.0     // 1 minute

const ServerIDKey = "serverID"
const ServerHostKey = "Host"
const TimeOutKey = "timeOut"
const DownloadSizesKey = "downloadSizes"
const UploadSizesKey = "uploadSizes"
const MaxSecondsKey = "maxSeconds"
const TestTypeKey = "testType"

func GetDefaultSpeedTestDownloadSizes() []int {
	return []int{245388, 505544, 1118012, 1986284}
}

func GetDefaultSpeedTestUploadSizes() []int {
	return []int{32768, 65536, 131072, 262144, 524288}
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	_, nodeSpecified := req.PathParameters["id"]
	switch req.HTTPMethod {
	case "DELETE":
		return deleteNode(req)
	case "GET":
		if nodeSpecified {
			if strings.HasSuffix(req.Path, "/tag") {
				return listNodeTags(req)
			}
			return viewNode(req)
		}
		return listNodes(req)
	case "PUT":
		return updateNode(req)
	default:
		return domain.ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func deleteNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []domain.Tag{})
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var node domain.Node
	err := db.DeleteItem(&node, id)
	return domain.ReturnJsonOrError(node, err)
}

func viewNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var node domain.Node
	err := db.GetItem(&node, id)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Ensure user is authorized ...
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	user, err := db.GetUserFromRequest(req)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	if !domain.CanUserUseNode(user, node) {
		return domain.ClientError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	return domain.ReturnJsonOrError(node, err)
}

func listNodeTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var node domain.Node
	err := db.GetItem(&node, id)
	return domain.ReturnJsonOrError(node.Tags, err)
}

func listNodes(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var allNodes []domain.Node
	err := db.ListItems(&allNodes, "nickname asc")
	if err != nil {
		return domain.ReturnJsonOrError([]domain.Node{}, err)
	}

	user, err := db.GetUserFromRequest(req)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	visibleNodes := []domain.Node{}
	for _, node := range allNodes {
		if domain.CanUserUseNode(user, node) {
			visibleNodes = append(visibleNodes, node)
		}
	}

	return domain.ReturnJsonOrError(visibleNodes, err)
}

// Nodes are only created through the agent /hello API and updated via admin /node APIs
func updateNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := domain.GetResourceIDFromRequest(req)
	if id == 0 {
		return domain.ClientError(http.StatusBadRequest, "Invalid ID")
	}

	var node domain.Node
	err := db.GetItem(&node, id)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// authorize request
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	// Get the node struct from the request body
	var updatedNode domain.Node
	err = json.Unmarshal([]byte(req.Body), &updatedNode)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Make sure tags are valid and user calling api is allowed to use them
	if !db.AreTagsValid(updatedNode.Tags) {
		return domain.ClientError(http.StatusBadRequest, "One or more submitted tags are invalid")
	}

	// check if user making api call can use the updated tags.
	statusCode, errMsg = db.GetAuthorizationStatus(req, domain.PermissionTagBased, updatedNode.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	cleanMac, err := domain.CleanMACAddress(updatedNode.MacAddr)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	if node.MacAddr == "" {
		node.MacAddr = cleanMac
	} else if node.MacAddr != updatedNode.MacAddr {
		err = fmt.Errorf("cannot change the mac address of an existing node")
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Apply updates to node
	node.Nickname = updatedNode.Nickname
	node.Notes = updatedNode.Notes

	node, err = updateNodeTasks(node)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	replaceAssoc := []domain.AssociationReplacement{
		{
			Replacement:     updatedNode.Tags,
			AssociationName: "Tags",
		},
		{
			Replacement:     updatedNode.Contacts,
			AssociationName: "Contacts",
		},
		{
			Replacement:     updatedNode.Tasks,
			AssociationName: "Tasks",
		},
		{
			Replacement:     updatedNode.ConfiguredVersion,
			AssociationName: "ConfiguredVersion",
		},
	}

	// Update the node in the database
	err = db.PutItemWithAssociations(&node, replaceAssoc)
	return domain.ReturnJsonOrError(node, err)
}

func updateNodeTasks(node domain.Node) (domain.Node, error) {
	newTasks := []domain.Task{}
	for index, task := range node.Tasks {
		if task.Type == domain.TaskTypePing {
			newTask, err := updateTaskPing(task)
			if err != nil {
				return node, fmt.Errorf("error updating task. Index: %d. Type: %s ... %s", index, task.Type, err.Error())
			}
			newTasks = append(newTasks, newTask)
		} else if task.Type == domain.TaskTypeSpeedTest {
			newTask, err := updateTaskSpeedTest(task)
			if err != nil {
				return node, fmt.Errorf("error updating task. Index: %d. Type: %s ... %s", index, task.Type, err.Error())
			}
			newTasks = append(newTasks, newTask)
		} else {
			newTasks = append(newTasks, task)
		}
	}

	node.Tasks = newTasks
	return node, nil
}

func updateTaskPing(task domain.Task) (domain.Task, error) {
	intValues := setIntValueIfMissing(task.TaskData.IntValues, TimeOutKey, DefaultPingTimeoutInSeconds)
	task.TaskData.IntValues = intValues

	stringValues, err := getPingStringValues(task)
	if err != nil {
		return task, err
	}
	task.TaskData.StringValues = stringValues

	return task, nil
}

func getPingStringValues(task domain.Task) (map[string]string, error) {
	stringValues := map[string]string{}
	if task.TaskData.StringValues != nil {
		stringValues = task.TaskData.StringValues
	}

	stringValues[TestTypeKey] = domain.TestConfigLatencyTest

	// If no NamedServerID, then use defaults
	if task.NamedServer.ID == 0 {
		stringValues = setStringValue(stringValues, ServerHostKey, domain.DefaultPingServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, domain.DefaultPingServerID)
		return stringValues, nil
	}

	// There is a NamedServerID but we're not checking if it's associated with a SpeedTestNetServer (for Pings)
	var namedServer domain.NamedServer
	err := db.GetItem(&namedServer, task.NamedServer.ID)
	if err != nil {
		return stringValues, fmt.Errorf("error getting NamedServer with UID: %d ... %s", task.NamedServer.ID, err.Error())
	}

	// If this does not refer to a SpeedTestNetServer, just use the NamedServer's values
	if namedServer.ServerType != domain.ServerTypeSpeedTestNet {
		stringValues = setStringValue(stringValues, ServerHostKey, namedServer.ServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, fmt.Sprintf("%v", namedServer.ID))
		return stringValues, nil
	}

	// This does refer to a SpeedTestNetServer, so use its info
	var speedTestNetServer domain.SpeedTestNetServer
	err = db.GetItem(&speedTestNetServer, namedServer.SpeedTestNetServerID)
	if err != nil {
		return stringValues, err
	}

	stringValues = setStringValue(stringValues, ServerHostKey, speedTestNetServer.Host)
	stringValues = setStringValue(stringValues, ServerIDKey, speedTestNetServer.ServerID)

	return stringValues, nil
}

func updateTaskSpeedTest(task domain.Task) (domain.Task, error) {
	intValues := setIntValueIfMissing(task.TaskData.IntValues, TimeOutKey, DefaultSpeedTestTimeoutInSeconds)
	task.TaskData.IntValues = intValues

	stringValues, err := getSpeedTestStringValues(task)
	if err != nil {
		return task, err
	}
	task.TaskData.StringValues = stringValues

	intSlices := setIntSliceIfMissing(task.TaskData.IntSlices, DownloadSizesKey, GetDefaultSpeedTestDownloadSizes())
	intSlices = setIntSliceIfMissing(intSlices, UploadSizesKey, GetDefaultSpeedTestUploadSizes())
	task.TaskData.IntSlices = intSlices

	floatValues := setFloatValueIfMissing(task.TaskData.FloatValues, MaxSecondsKey, DefaultSpeedTestMaxSeconds)
	task.TaskData.FloatValues = floatValues

	return task, nil
}

func getSpeedTestStringValues(task domain.Task) (map[string]string, error) {
	stringValues := map[string]string{}
	if task.TaskData.StringValues != nil {
		stringValues = task.TaskData.StringValues
	}

	stringValues[TestTypeKey] = domain.TestConfigSpeedTest

	// If there is no NamedServerID, then use the defaults
	if task.NamedServer.ID == 0 {
		stringValues = setStringValue(stringValues, ServerHostKey, domain.DefaultSpeedTestNetServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, domain.DefaultSpeedTestNetServerID)
		return stringValues, nil
	}

	// There is a NamedServerID
	var namedServer domain.NamedServer
	err := db.GetItem(&namedServer, task.NamedServer.ID)
	if err != nil {
		return stringValues, fmt.Errorf("error getting NamedServer with ID: %v ... %s", task.NamedServer.ID, err.Error())
	}

	// If this does not refer to a SpeedTestNetServer, just use the NamedServer's values
	if namedServer.ServerType != domain.ServerTypeSpeedTestNet {
		stringValues = setStringValue(stringValues, ServerHostKey, namedServer.ServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, fmt.Sprintf("%v", namedServer.ID))
		return stringValues, nil
	}

	// This does refer to a SpeedTestNetServer, so use its info
	var speedTestNetServer domain.SpeedTestNetServer
	err = db.GetItem(&speedTestNetServer, namedServer.SpeedTestNetServerID)
	if err != nil {
		return stringValues, err
	}

	stringValues = setStringValue(stringValues, ServerHostKey, speedTestNetServer.Host)
	stringValues = setStringValue(stringValues, ServerIDKey, speedTestNetServer.ServerID)
	return stringValues, nil
}

func setIntValueIfMissing(intValues map[string]int, key string, value int) map[string]int {
	if intValues == nil {
		return map[string]int{key: value}
	}

	_, ok := intValues[key]
	if !ok {
		intValues[key] = value
	}
	return intValues
}

func setFloatValueIfMissing(floatValues map[string]float64, key string, value float64) map[string]float64 {
	if floatValues == nil {
		return map[string]float64{key: value}
	}

	_, ok := floatValues[key]
	if !ok {
		floatValues[key] = value
	}
	return floatValues
}

func setIntSliceIfMissing(intSlices map[string][]int, key string, values []int) map[string][]int {
	if intSlices == nil {
		return map[string][]int{key: values}
	}

	_, ok := intSlices[key]
	if !ok {
		intSlices[key] = values
	}
	return intSlices
}

func setStringValue(stringValues map[string]string, key string, value string) map[string]string {
	if stringValues == nil {
		return map[string]string{key: value}
	}

	stringValues[key] = value
	return stringValues
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

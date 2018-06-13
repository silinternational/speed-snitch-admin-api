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
	_, nodeSpecified := req.PathParameters["macAddr"]
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

	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])

	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	success, err := db.DeleteItem(domain.DataTable, SelfType, macAddr)

	if err != nil {
		return domain.ServerError(err)
	}

	if !success {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func viewNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	node, err := db.GetNode(macAddr)
	if err != nil {
		return domain.ServerError(err)
	}

	if node.Arch == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	// Ensure user is authorized ...
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.Tags)
	if statusCode > 0 {
		return domain.ClientError(statusCode, errMsg)
	}

	js, err := json.Marshal(node)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listNodeTags(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	node, err := db.GetNode(macAddr)
	if err != nil {
		return domain.ServerError(err)
	}

	js, err := json.Marshal(node.Tags)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func listNodes(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	nodes, err := db.ListNodes()
	if err != nil {
		return domain.ServerError(err)
	}

	user, err := db.GetUserFromRequest(req)
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}

	var jsBody string

	if user.Role == domain.UserRoleSuperAdmin {
		jsBody, err = domain.GetJSONFromSlice(nodes)
		if err != nil {
			return domain.ServerError(err)
		}
	} else {
		visibleNodes := []domain.Node{}
		for _, node := range nodes {
			if domain.CanUserUseNode(user, node) {
				visibleNodes = append(visibleNodes, node)
			}
		}

		jsBody, err = domain.GetJSONFromSlice(visibleNodes)
		if err != nil {
			return domain.ServerError(err)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       jsBody,
	}, nil
}

func updateNode(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var node domain.Node

	if req.PathParameters["macAddr"] == "" {
		return domain.ClientError(http.StatusBadRequest, "Mac Address is required")
	}

	// Clean the MAC Address
	macAddr, err := domain.CleanMACAddress(req.PathParameters["macAddr"])
	if err != nil {
		return domain.ClientError(http.StatusBadRequest, err.Error())
	}
	node, err = db.GetNode(macAddr)
	if err != nil {
		return domain.ServerError(err)
	}

	// Get the node struct from the request body
	var updatedNode domain.Node
	err = json.Unmarshal([]byte(req.Body), &updatedNode)
	if err != nil {
		return domain.ServerError(err)
	}

	// Make sure tags are valid and user calling api is allowed to use them
	if !db.AreTagsValid(updatedNode.Tags) {
		return domain.ClientError(http.StatusBadRequest, "One or more submitted tags are invalid")
	}
	// @todo do we need to check if user making api call can use the tags provided?

	// Apply updates to node
	node.ID = SelfType + "-" + macAddr
	node.ConfiguredVersion = updatedNode.ConfiguredVersion
	node.Tasks = updatedNode.Tasks
	node.Contacts = updatedNode.Contacts
	node.Tags = updatedNode.Tags
	node.Nickname = updatedNode.Nickname
	node.Notes = updatedNode.Notes

	// If node already exists, ensure user is authorized ...
	existingNode, err := db.GetNode(macAddr)
	if err == nil {
		statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, existingNode.Tags)
		if statusCode > 0 {
			return domain.ClientError(statusCode, errMsg)
		}
	}

	// We need to use Client{} to allow for unit testing the function
	// It just calls the common.db methods with the same names
	node, err = updateNodeTasks(node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Update the node in the database
	err = db.PutItem(domain.DataTable, node)
	if err != nil {
		return domain.ServerError(err)
	}

	// Return the updated node as json
	js, err := json.Marshal(node)
	if err != nil {
		return domain.ServerError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func updateNodeTasks(node domain.Node) (domain.Node, error) {
	newTasks := []domain.Task{}
	for index, task := range node.Tasks {
		if task.Type == domain.TaskTypePing {
			newTask, err := updateTaskPing(task)
			if err != nil {
				return node, fmt.Errorf("Error updating task. Index: %d. Type: %s ... %s", index, task.Type, err.Error())
			}
			newTasks = append(newTasks, newTask)
		} else if task.Type == domain.TaskTypeSpeedTest {
			newTask, err := updateTaskSpeedTest(task)
			if err != nil {
				return node, fmt.Errorf("Error updating task. Index: %d. Type: %s ... %s", index, task.Type, err.Error())
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
	intValues := setIntValueIfMissing(task.Data.IntValues, TimeOutKey, DefaultPingTimeoutInSeconds)
	task.Data.IntValues = intValues

	stringValues, err := getPingStringValues(task)
	if err != nil {
		return task, err
	}
	task.Data.StringValues = stringValues

	return task, nil
}

func getPingStringValues(task domain.Task) (map[string]string, error) {
	stringValues := map[string]string{}
	if task.Data.StringValues != nil {
		stringValues = task.Data.StringValues
	}

	stringValues[TestTypeKey] = domain.TestConfigLatencyTest

	// If no NamedServerID, then use defaults
	if task.NamedServer.ID == "" {
		stringValues = setStringValue(stringValues, ServerHostKey, domain.DefaultPingServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, domain.DefaultPingServerID)
		return stringValues, nil
	}

	// There is a NamedServerID but we're not checking if it's associated with a SpeedTestNetServer (for Pings)
	var namedServer domain.NamedServer
	namedServer, err := db.GetNamedServer(task.NamedServer.UID)
	if err != nil {
		return stringValues, fmt.Errorf("Error getting NamedServer with UID: %s ... %s", task.NamedServer.ID, err.Error())
	}

	stringValues = setStringValue(stringValues, ServerHostKey, namedServer.ServerHost)
	stringValues = setStringValue(stringValues, ServerIDKey, namedServer.UID)

	return stringValues, nil
}

func updateTaskSpeedTest(task domain.Task) (domain.Task, error) {
	intValues := setIntValueIfMissing(task.Data.IntValues, TimeOutKey, DefaultSpeedTestTimeoutInSeconds)
	task.Data.IntValues = intValues

	stringValues, err := getSpeedTestStringValues(task)
	if err != nil {
		return task, err
	}
	task.Data.StringValues = stringValues

	intSlices := setIntSliceIfMissing(task.Data.IntSlices, DownloadSizesKey, GetDefaultSpeedTestDownloadSizes())
	intSlices = setIntSliceIfMissing(intSlices, UploadSizesKey, GetDefaultSpeedTestUploadSizes())
	task.Data.IntSlices = intSlices

	floatValues := setFloatValueIfMissing(task.Data.FloatValues, MaxSecondsKey, DefaultSpeedTestMaxSeconds)
	task.Data.FloatValues = floatValues

	return task, nil
}

func getSpeedTestStringValues(task domain.Task) (map[string]string, error) {
	stringValues := map[string]string{}
	if task.Data.StringValues != nil {
		stringValues = task.Data.StringValues
	}

	stringValues[TestTypeKey] = domain.TestConfigSpeedTest

	// If there is no NamedServerID, then use the defaults
	if task.NamedServer.ID == "" {
		stringValues = setStringValue(stringValues, ServerHostKey, domain.DefaultSpeedTestNetServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, domain.DefaultSpeedTestNetServerID)
		return stringValues, nil
	}

	// There is a NamedServerID
	namedServer, err := db.GetNamedServer(task.NamedServer.UID)
	if err != nil {
		return stringValues, fmt.Errorf("Error getting NamedServer with UID: %s ... %s", task.NamedServer.UID, err.Error())
	}

	// This does not refer to a SpeedTestNetServer
	if namedServer.ServerType != domain.ServerTypeSpeedTestNet {
		stringValues = setStringValue(stringValues, ServerHostKey, namedServer.ServerHost)
		stringValues = setStringValue(stringValues, ServerIDKey, namedServer.UID)
		return stringValues, nil
	}

	// This does refer to a SpeedTestNetServer, so use its info
	stnServer, err := db.GetSpeedTestNetServerFromNamedServer(namedServer)
	if err != nil {
		return stringValues, err
	}

	stringValues = setStringValue(stringValues, ServerHostKey, stnServer.Host)
	stringValues = setStringValue(stringValues, ServerIDKey, stnServer.ServerID)
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
	lambda.Start(router)
}

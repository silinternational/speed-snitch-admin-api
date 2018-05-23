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
const DefaultSpeedTestTimeoutInSeconds = 300 // 5 minutes
const DefaultSpeedTestMaxSeconds = 300.0     // 5 minutes

// This is need for testing
type dbClient interface {
	GetItem(string, string, string, interface{}) error
	GetSpeedTestNetServerFromNamedServer(domain.NamedServer) (domain.SpeedTestNetServer, error)
}

// This is needed to allow for mock db result in the tests
type Client struct{}

func (c Client) GetItem(tableAlias, dataType, value string, itemObj interface{}) error {
	return db.GetItem(tableAlias, dataType, value, itemObj)
}

func (c Client) GetSpeedTestNetServerFromNamedServer(namedServer domain.NamedServer) (domain.SpeedTestNetServer, error) {
	return db.GetSpeedTestNetServerFromNamedServer(namedServer)
}

func GetDefaultSpeedTestDownloadSizes() []int {
	return []int{245388, 505544}
}

func GetDefaultSpeedTestUploadSizes() []int {
	return []int{32768, 65536}
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
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionSuperAdmin, []string{})
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

	var node domain.Node
	err = db.GetItem(domain.DataTable, SelfType, macAddr, &node)
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
	statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, node.TagUIDs)
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

	var node domain.Node
	err = db.GetItem(domain.DataTable, SelfType, macAddr, &node)
	if err != nil {
		return domain.ServerError(err)
	}

	allTags, err := db.ListTags()
	if err != nil {
		return domain.ServerError(err)
	}

	var nodeTags []domain.Tag

	for _, tag := range allTags {
		inArray, _ := domain.InArray(tag.UID, node.TagUIDs)
		if inArray {
			nodeTags = append(nodeTags, tag)
		}
	}

	js, err := json.Marshal(nodeTags)
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
	err = db.GetItem(domain.DataTable, SelfType, macAddr, &node)
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
	if !db.AreTagsValid(updatedNode.TagUIDs) {
		return domain.ClientError(http.StatusBadRequest, "One or more submitted tags are invalid")
	}
	// @todo do we need to check if user making api call can use the tags provided?

	// Apply updates to node
	node.ID = SelfType + "-" + macAddr
	node.ConfiguredVersion = updatedNode.ConfiguredVersion
	node.Tasks = updatedNode.Tasks
	node.Contacts = updatedNode.Contacts
	node.TagUIDs = updatedNode.TagUIDs
	node.Nickname = updatedNode.Nickname
	node.Notes = updatedNode.Notes

	// If node already exists, ensure user is authorized ...
	var existingNode domain.Node
	err = db.GetItem(domain.DataTable, SelfType, macAddr, &existingNode)
	if err == nil {
		statusCode, errMsg := db.GetAuthorizationStatus(req, domain.PermissionTagBased, existingNode.TagUIDs)
		if statusCode > 0 {
			return domain.ClientError(statusCode, errMsg)
		}
	}

	// We need to use Client{} to allow for unit testing the function
	// It just calls the common.db methods with the same names
	node, err = updateNodeTasks(node, Client{})
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

func updateNodeTasks(node domain.Node, db dbClient) (domain.Node, error) {
	newTasks := []domain.Task{}
	for index, task := range node.Tasks {
		if task.Type == domain.TaskTypePing {
			newTask, err := updateTaskPing(task, db)
			if err != nil {
				return node, fmt.Errorf("Error updating task. Index: %d. Type: %s ... %s", index, task.Type, err.Error())
			}
			newTasks = append(newTasks, newTask)
		} else if task.Type == domain.TaskTypeSpeedTest {
			newTask, err := updateTaskSpeedTest(task, db)
			if err != nil {
				return node, fmt.Errorf("Error updating task. Index: %d. Type: %s ... %s", index, task.Type, err.Error())
			}
			newTasks = append(newTasks, newTask)
		} else {
			newTasks = append(newTasks, task)
		}
	}
	return node, nil
}

func updateTaskPing(task domain.Task, db dbClient) (domain.Task, error) {
	intValues := map[string]int{}
	if task.Data.IntValues != nil {
		intValues = task.Data.IntValues
	}

	intValues = setIntValueIfMissing(intValues, "timeOut", DefaultPingTimeoutInSeconds)
	task.Data.IntValues = intValues

	stringValues, err := getPingStringValues(task, db)
	if err != nil {
		return task, err
	}
	task.Data.StringValues = stringValues

	return task, nil
}

func getPingStringValues(task domain.Task, db dbClient) (map[string]string, error) {
	stringValues := map[string]string{}
	if task.Data.StringValues != nil {
		stringValues = task.Data.StringValues
	}

	stringValues["testType"] = domain.TestConfigLatencyTest

	// If no NamedServerID, then use defaults
	if task.NamedServerID == "" {
		stringValues = setStringValueIfMissing(stringValues, "Host", domain.DefaultPingServerHost)
		stringValues = setStringValueIfMissing(stringValues, "serverID", domain.DefaultPingServerID)
		return stringValues, nil
	}

	// There is a NamedServerID but we're not checking if it's associated with a SpeedTestNetServer (for Pings)
	var namedServer domain.NamedServer
	err := db.GetItem(domain.DataTable, domain.DataTypeNamedServer, task.NamedServerID, &namedServer)
	if err != nil {
		return stringValues, fmt.Errorf("Error getting NamedServer with UID: %s ... %s", task.NamedServerID, err.Error())
	}

	stringValues = setStringValueIfMissing(stringValues, "Host", namedServer.ServerHost)
	stringValues = setStringValueIfMissing(stringValues, "serverID", namedServer.UID)

	return stringValues, nil
}

func updateTaskSpeedTest(task domain.Task, db dbClient) (domain.Task, error) {
	intValues := map[string]int{}
	if task.Data.IntValues != nil {
		intValues = task.Data.IntValues
	}

	intValues = setIntValueIfMissing(intValues, "timeOut", DefaultSpeedTestTimeoutInSeconds)
	task.Data.IntValues = intValues

	stringValues, err := getSpeedTestStringValues(task, db)
	if err != nil {
		return task, err
	}
	task.Data.StringValues = stringValues

	intSlices := map[string][]int{}
	if task.Data.IntSlices != nil {
		intSlices = task.Data.IntSlices
	}
	intSlices = setIntSliceIfMissing(intSlices, "downloadSizes", GetDefaultSpeedTestDownloadSizes())
	intSlices = setIntSliceIfMissing(intSlices, "uploadSizes", GetDefaultSpeedTestUploadSizes())
	task.Data.IntSlices = intSlices

	floatValues := map[string]float64{}
	if task.Data.FloatValues != nil {
		floatValues = task.Data.FloatValues
	}
	floatValues = setFloatValueIfMissing(floatValues, "maxSeconds", DefaultSpeedTestMaxSeconds)
	task.Data.FloatValues = floatValues

	return task, nil
}

func getSpeedTestStringValues(task domain.Task, db dbClient) (map[string]string, error) {
	stringValues := map[string]string{}
	if task.Data.StringValues != nil {
		stringValues = task.Data.StringValues
	}

	stringValues["testType"] = domain.TestConfigSpeedTest

	// If there is no NamedServerID, then use the defaults
	if task.NamedServerID == "" {
		stringValues = setStringValueIfMissing(stringValues, "Host", domain.DefaultSpeedTestNetServerHost)
		stringValues = setStringValueIfMissing(stringValues, "serverID", domain.DefaultSpeedTestNetServerID)
		return stringValues, nil
	}

	// There is a NamedServerID
	var namedServer domain.NamedServer
	err := db.GetItem(domain.DataTable, domain.DataTypeNamedServer, task.NamedServerID, &namedServer)
	if err != nil {
		return stringValues, fmt.Errorf("Error getting NamedServer with UID: %s ... %s", task.NamedServerID, err.Error())
	}

	// This does not refer to a SpeedTestNetServer
	if namedServer.ServerType != domain.ServerTypeSpeedTestNet {
		stringValues = setStringValueIfMissing(stringValues, "Host", namedServer.ServerHost)
		stringValues = setStringValueIfMissing(stringValues, "serverID", namedServer.UID)
		return stringValues, nil
	}

	// This does refer to a SpeedTestNetServer, so use its info
	stnServer, err := db.GetSpeedTestNetServerFromNamedServer(namedServer)
	if err != nil {
		return stringValues, err
	}

	stringValues = setStringValueIfMissing(stringValues, "Host", stnServer.Host)
	stringValues = setStringValueIfMissing(stringValues, "serverID", stnServer.ServerID)
	return stringValues, nil
}

func setIntValueIfMissing(intValues map[string]int, key string, value int) map[string]int {
	_, ok := intValues[key]
	if !ok {
		intValues[key] = value
	}
	return intValues
}

func setFloatValueIfMissing(floatValues map[string]float64, key string, value float64) map[string]float64 {
	_, ok := floatValues[key]
	if !ok {
		floatValues[key] = value
	}
	return floatValues
}

func setIntSliceIfMissing(intSlices map[string][]int, key string, values []int) map[string][]int {
	_, ok := intSlices[key]
	if !ok {
		intSlices[key] = values
	}
	return intSlices
}

func setStringValueIfMissing(stringValues map[string]string, key string, value string) map[string]string {
	_, ok := stringValues[key]
	if !ok {
		stringValues[key] = value
	}
	return stringValues
}

func main() {
	lambda.Start(router)
}

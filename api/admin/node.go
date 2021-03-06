package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

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

func nodeRouter(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	if node.Tags == nil {
		node.Tags = []domain.Tag{}
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
	if node.Tags == nil {
		node.Tags = []domain.Tag{}
	}
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

	businessStartTime, businessCloseTime, err := domain.CleanBusinessTimes(
		updatedNode.BusinessStartTime,
		updatedNode.BusinessCloseTime,
	)

	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Apply updates to node
	node.BusinessStartTime = businessStartTime
	node.BusinessCloseTime = businessCloseTime
	node.Nickname = updatedNode.Nickname
	node.Notes = updatedNode.Notes

	updatedNode, err = updateNodeTasks(updatedNode)
	if err != nil {
		return domain.ReturnJsonOrError(domain.Node{}, err)
	}

	// Get new node version
	if updatedNode.ConfiguredVersionID > 0 && updatedNode.ConfiguredVersionID != node.ConfiguredVersion.ID {
		var newVersion domain.Version
		err := db.GetItem(&newVersion, updatedNode.ConfiguredVersionID)
		if err != nil {
			err = fmt.Errorf(
				"error getting updated configured version with ID: %d\n%s",
				updatedNode.ConfiguredVersionID,
				err.Error(),
			)
			return domain.ReturnJsonOrError(domain.Node{}, err)
		}

		updatedNode.ConfiguredVersion = newVersion
	}

	replaceAssoc := []domain.AssociationReplacements{
		{
			Replacements:    updatedNode.Tags,
			AssociationName: "Tags",
		},
		{
			Replacements:    updatedNode.Contacts,
			AssociationName: "Contacts",
		},
		{
			Replacements:    updatedNode.Tasks,
			AssociationName: "Tasks",
		},
		{
			Replacements:    []domain.Version{updatedNode.ConfiguredVersion},
			AssociationName: "ConfiguredVersion",
		},
	}

	// Update the node in the database
	err = db.PutItemWithAssociations(&node, replaceAssoc)
	if node.Tags == nil {
		node.Tags = []domain.Tag{}
	}

	task := domain.Task{}
	db.DeleteOrphanedItems(&task, "node_id")

	contact := domain.Contact{}
	db.DeleteOrphanedItems(&contact, "node_id")

	return domain.ReturnJsonOrError(node, err)
}

func updateNodeTasks(node domain.Node) (domain.Node, error) {
	newTasks := []domain.Task{}
	for index, task := range node.Tasks {
		if task.Type == domain.TaskTypeSpeedTest || task.Type == domain.TaskTypePing {
			if task.NamedServerID == 0 {
				err := fmt.Errorf("task of type %s must have a NamedServerID.", task.Type)
				return node, err
			}
			var namedServer domain.NamedServer
			err := db.GetItem(&namedServer, task.NamedServerID)
			if err != nil {
				return domain.Node{}, err
			}
			task.NamedServer = namedServer
			task.ServerHost = namedServer.ServerHost
		}
		if task.Type == domain.TaskTypeSpeedTest {
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

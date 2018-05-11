package db

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fillup/semver"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-agent"
	"github.com/silinternational/speed-snitch-agent/lib/speedtestnet"
	"net/http"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

func GetItem(tableAlias, dataType, value string, itemObj interface{}) error {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(domain.GetDbTableName(tableAlias)),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(dataType + "-" + value),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return err
	}
	if result.Item == nil {
		return nil
	}

	// The result.Item object returned has the underlying type
	// map[string]*AttributeValue. We can use the UnmarshalMap helper
	// to parse this straight into the fields of a struct. Note:
	// UnmarshalListOfMaps also exists if you are working with multiple
	// items.
	err = dynamodbattribute.UnmarshalMap(result.Item, itemObj)
	if err != nil {
		return err
	}

	return nil
}

func PutItem(tableAlias string, item interface{}) error {
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		domain.ServerError(fmt.Errorf("failed to DynamoDB marshal Record, %v", err))
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(domain.GetDbTableName(tableAlias)),
		Item:      av,
	}

	_, err = db.PutItem(input)
	return err
}

func DeleteItem(tableAlias, dataType, value string) (bool, error) {

	// Prepare the input for the query.
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(domain.GetDbTableName(tableAlias)),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(dataType + "-" + value),
			},
		},
	}

	// Delete the item from DynamoDB. I
	_, err := db.DeleteItem(input)

	if err != nil && err.Error() == dynamodb.ErrCodeReplicaNotFoundException {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func scanTable(tableAlias, dataType string) ([]map[string]*dynamodb.AttributeValue, error) {
	tableName := domain.GetDbTableName(tableAlias)
	filterExpression := "begins_with(ID, :dataType)"
	input := &dynamodb.ScanInput{
		TableName:        &tableName,
		FilterExpression: &filterExpression,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":dataType": {
				S: aws.String(dataType + "-"),
			},
		},
	}

	var results []map[string]*dynamodb.AttributeValue
	err := db.ScanPages(input,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			results = append(results, page.Items...)
			return !lastPage
		})

	if err != nil {
		return results, err
	}

	return results, nil
}

func GetUserByUserID(userID string) (domain.User, error) {
	tableName := domain.GetDbTableName(domain.DataTable)
	filterExpression := "UserID = :userID"
	input := &dynamodb.ScanInput{
		TableName:        &tableName,
		FilterExpression: &filterExpression,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userID": {
				S: aws.String(userID),
			},
		},
	}

	var results []map[string]*dynamodb.AttributeValue
	err := db.ScanPages(input,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			results = append(results, page.Items...)
			return !lastPage
		})
	if err != nil {
		return domain.User{}, err
	}

	if len(results) == 0 {
		return domain.User{}, nil
	}
	if len(results) == 1 {
		var user domain.User
		err := dynamodbattribute.UnmarshalMap(results[0], &user)
		if err != nil {
			return domain.User{}, err
		}

		return user, nil
	}

	return domain.User{}, fmt.Errorf("More than one user found for UserID %s", userID)
}

func ListTags() ([]domain.Tag, error) {

	var list []domain.Tag

	items, err := scanTable(domain.DataTable, "tag")
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.Tag
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.Tag{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

func ListNodes() ([]domain.Node, error) {

	var list []domain.Node

	items, err := scanTable(domain.DataTable, "node")
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.Node
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.Node{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

func ListVersions() ([]domain.Version, error) {

	var list []domain.Version

	items, err := scanTable(domain.DataTable, "version")
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.Version
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.Version{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

func ListUsers() ([]domain.User, error) {

	var list []domain.User

	items, err := scanTable(domain.DataTable, "user")
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.User
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.User{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

func ListSpeedTestNetServers() ([]domain.SpeedTestNetServer, error) {

	var list []domain.SpeedTestNetServer

	items, err := scanTable(domain.DataTable, "speedtestnetserver")
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.SpeedTestNetServer
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.SpeedTestNetServer{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

type ServerData struct {
	ID   int
	Host string
}

type NodesForServer struct {
	ServerData ServerData
	Nodes      []domain.Node
}

func GetServerDataFromNode(node domain.Node) ([]ServerData, error) {
	// Use a map to avoid multiple entries for the same server
	tempData := map[int]ServerData{}

	for _, task := range node.Tasks {
		if task.Type != agent.TypePing && task.Type != agent.TypeSpeedTest {
			continue
		}

		id, ok := task.Data.IntValues[speedtestnet.CFG_SERVER_ID]
		if !ok {
			err := fmt.Errorf("task.Data.IntValues is missing an entry for %s", speedtestnet.CFG_SERVER_ID)
			return []ServerData{}, err
		}

		host, ok := task.Data.StringValues[speedtestnet.CFG_SERVER_HOST]
		if !ok {
			err := fmt.Errorf("task.Data.StringValues is missing an entry for %s", speedtestnet.CFG_SERVER_HOST)
			return []ServerData{}, err
		}

		tempData[id] = ServerData{ID: id, Host: host}
	}

	// Convert the map back to a slice
	allServerData := []ServerData{}

	for _, data := range tempData {
		allServerData = append(allServerData, data)
	}

	return allServerData, nil
}

// GetNodesForServers returns something like this ...
//   {<serverID> : {{ID: <server ID>, Host: "<server Host>"}, [<node1>, <node2>]}, ...  }
//   In other words, each speedtest server has an entry in the output map that includes a
//     a struct with its ID and Host and a list of the nodes that are meant to use it.
func GetNodesForServers(nodes []domain.Node) (map[int]NodesForServer, error) {
	allNodesForServer := map[int]NodesForServer{}

	// For each node in the database, add its servers to the output
	for _, node := range nodes {
		serverDataList, err := GetServerDataFromNode(node)

		if err != nil {
			err = fmt.Errorf("Error getting server data for node with MAC Address %s\n\t%s", node.MacAddr, err.Error())
			return map[int]NodesForServer{}, err
		}

		// For each server, make sure it is in the output and add the current node to its list of nodes
		for _, serverData := range serverDataList {
			// If this server isn't in the output, include it
			nodesForServer, ok := allNodesForServer[serverData.ID]
			if !ok {
				nodesForServer = NodesForServer{ServerData: serverData}
			}

			// Add this node to this servers list of nodes
			nodesForServer.Nodes = append(nodesForServer.Nodes, node)
			allNodesForServer[serverData.ID] = nodesForServer
		}
	}

	return allNodesForServer, nil
}

// GetLatestVersion iterates through version in db to return only the latest version
func GetLatestVersion() (domain.Version, error) {
	versions, err := ListVersions()
	if err != nil {
		return domain.Version{}, err
	}

	var latest domain.Version

	for _, version := range versions {
		if latest.Number == "" {
			latest = version
		} else {
			isNewer, err := semver.IsNewer(latest.Number, version.Number)
			if err != nil {
				return domain.Version{}, err
			}
			if isNewer {
				latest = version
			}
		}

	}

	return latest, nil
}

func GetUserFromRequest(req events.APIGatewayProxyRequest) (domain.User, error) {
	userID, ok := req.Headers[domain.UserReqHeaderID]
	if !ok {
		return domain.User{}, fmt.Errorf("Missing Header: %s", domain.UserReqHeaderID)
	}

	// Get the user
	user, err := GetUserByUserID(userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// GetAuthorizationStatus returns 0, nil for users that are authorized to use the object
func GetAuthorizationStatus(req events.APIGatewayProxyRequest, permissionType string, objectTagUIDs []string) (int, string) {
	user, err := GetUserFromRequest(req)
	if err != nil {
		return http.StatusBadRequest, err.Error()
	}

	if user.Role == domain.PermissionSuperAdmin {
		return 0, ""
	}

	if permissionType == domain.PermissionSuperAdmin {
		return http.StatusForbidden, http.StatusText(http.StatusForbidden)
	}

	if permissionType == domain.PermissionTagBased {
		tagsOverlap := domain.DoTagsOverlap(user.TagUIDs, objectTagUIDs)
		if tagsOverlap {
			return 0, ""
		}

		return http.StatusForbidden, http.StatusText(http.StatusForbidden)
	}

	return http.StatusInternalServerError, "Invalid permission type requested: " + permissionType
}

func AreTagsValid(tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	allTags, err := ListTags()
	if err != nil {
		return false
	}

	allTagUIDs := []string{}
	for _, tag := range allTags {
		allTagUIDs = append(allTagUIDs, tag.UID)
	}

	for _, tag := range tags {
		inArray, _ := domain.InArray(tag, allTagUIDs)
		if !inArray {
			return false
		}
	}

	return true
}

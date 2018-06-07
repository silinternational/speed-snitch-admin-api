package db

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/fillup/semver"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-agent"
	"github.com/silinternational/speed-snitch-agent/lib/speedtestnet"
	"net/http"
	"os"
)

const ENV_DYNAMO_ENDPOINT = "AWS_DYNAMODB_ENDPOINT"

var db *dynamodb.DynamoDB

func GetDb() *dynamodb.DynamoDB {
	if db == nil {
		dynamoEndpoint := os.Getenv(ENV_DYNAMO_ENDPOINT)
		fmt.Fprintf(os.Stdout, "dynamodb endpoint: %s\n", dynamoEndpoint)
		db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1").WithEndpoint(dynamoEndpoint))
	}
	return db
}

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

	db := GetDb()

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

	db := GetDb()
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

	db := GetDb()
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

func ScanTable(tableAlias, dataType string) ([]map[string]*dynamodb.AttributeValue, error) {
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

	db := GetDb()
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

func scanTaskLogForRange(startTime, endTime int64, nodeMacAddr string, logTypeIDPrefixes []string) ([]map[string]*dynamodb.AttributeValue, error) {
	tableName := domain.GetDbTableName(domain.TaskLogTable)

	var queryCondition expression.ConditionBuilder

	// startTime and endTime are required, so start condition with it
	timeCondition := expression.Between(expression.Name("Timestamp"), expression.Value(startTime), expression.Value(endTime))

	// Default queryCondition is time alone
	queryCondition = timeCondition

	// If a list of ID prefixes were provided, append to query condition
	//if len(logTypeIDPrefixes) > 0 {
	//	conditionOrs := []expression.ConditionBuilder{}
	//	for _, prefix := range logTypeIDPrefixes {
	//		newOr := expression.BeginsWith(expression.Name("ID"), prefix)
	//		conditionOrs = append(conditionOrs, newOr)
	//	}
	//	if len(logTypeIDPrefixes) == 1 {
	//		queryCondition.And(timeCondition, conditionOrs[0])
	//	} else if len(logTypeIDPrefixes) == 2 {
	//		queryCondition = expression.And(timeCondition, conditionOrs[0], conditionOrs[1])
	//	} else {
	//		queryCondition.And(timeCondition, expression.Or(conditionOrs[0], conditionOrs[1], conditionOrs...))
	//	}
	//} else {
	//	queryCondition = timeCondition
	//}

	prefixCount := len(logTypeIDPrefixes)
	if prefixCount > 0 {
		if prefixCount == 1 {
			queryCondition = expression.And(timeCondition, expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[0]))
		} else if prefixCount == 2 {
			orCondition := expression.Or(
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[0]),
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[1]))
			queryCondition = expression.And(timeCondition, orCondition)
		} else if prefixCount == 3 {
			orCondition := expression.Or(
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[0]),
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[1]),
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[2]))
			queryCondition = expression.And(timeCondition, orCondition)
		} else if prefixCount == 4 {
			orCondition := expression.Or(
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[0]),
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[1]),
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[2]),
				expression.BeginsWith(expression.Name("ID"), logTypeIDPrefixes[3]))
			queryCondition = expression.And(timeCondition, orCondition)
		}
	}

	// conditional macAddr condition
	if nodeMacAddr != "" {
		macAddrCondition := expression.Contains(expression.Name("MacAddr"), nodeMacAddr)
		queryCondition = expression.And(queryCondition, macAddrCondition)
	}

	expr, err := expression.NewBuilder().WithCondition(queryCondition).Build()
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	consistentRead := true
	input := &dynamodb.ScanInput{
		TableName:                 &tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Condition(),
		ConsistentRead:            &consistentRead,
	}

	db := GetDb()
	var results []map[string]*dynamodb.AttributeValue
	err = db.ScanPages(input,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			results = append(results, page.Items...)
			return !lastPage
		})

	if err != nil {
		return results, err
	}

	return results, nil
}

func GetTaskLogForRange(startTime, endTime int64, nodeMacAddr string, logTypeIDPrefixes []string) ([]domain.TaskLogEntry, error) {
	var results []domain.TaskLogEntry

	items, err := scanTaskLogForRange(startTime, endTime, nodeMacAddr, logTypeIDPrefixes)
	if err != nil {
		return results, err
	}

	for _, item := range items {
		var itemObj domain.TaskLogEntry
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.TaskLogEntry{}, err
		}
		results = append(results, itemObj)
	}

	return results, nil
}

func GetDailySnapshotsForRange(startTime, endTime int64, nodeMacAddr string) ([]domain.DailySnapshot, error) {
	var results []domain.DailySnapshot

	items, err := scanTaskLogForRange(startTime, endTime, nodeMacAddr, []string{"daily"})
	if err != nil {
		return results, err
	}

	for _, item := range items {
		var itemObj domain.DailySnapshot
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.DailySnapshot{}, err
		}
		results = append(results, itemObj)
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
	db := GetDb()
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

	items, err := ScanTable(domain.DataTable, "tag")
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

	items, err := ScanTable(domain.DataTable, "node")
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

	items, err := ScanTable(domain.DataTable, "version")
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

	items, err := ScanTable(domain.DataTable, "user")
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

func ListNamedServers() ([]domain.NamedServer, error) {

	var list []domain.NamedServer

	items, err := ScanTable(domain.DataTable, domain.DataTypeNamedServer)
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.NamedServer
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.NamedServer{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

// ListSTNetServerLists returns all the country-grouped rows of
//  speedtest.net servers
func ListSTNetServerLists() ([]domain.STNetServerList, error) {

	var list []domain.STNetServerList

	items, err := ScanTable(domain.DataTable, domain.DataTypeSTNetServerList)
	if err != nil {
		return list, err
	}

	for _, item := range items {
		var itemObj domain.STNetServerList
		err := dynamodbattribute.UnmarshalMap(item, &itemObj)
		if err != nil {
			return []domain.STNetServerList{}, err
		}
		list = append(list, itemObj)
	}

	return list, nil
}

func GetSpeedTestNetServerFromNamedServer(namedServer domain.NamedServer) (domain.SpeedTestNetServer, error) {
	var stnServerList domain.STNetServerList
	countryCode := namedServer.Country.Code
	err := GetItem(domain.DataTable, domain.DataTypeSTNetServerList, countryCode, &stnServerList)
	if err != nil {
		return domain.SpeedTestNetServer{}, fmt.Errorf("Error getting STNetServerList for NamedServer with UID: %s ... %s", namedServer.UID, err.Error())
	}

	for _, server := range stnServerList.Servers {
		if server.ServerID == namedServer.SpeedTestNetServerID {
			return server, nil
		}
	}

	return domain.SpeedTestNetServer{}, fmt.Errorf("Could not find matching SpeedTestNet Server with id %s.", namedServer.SpeedTestNetServerID)
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

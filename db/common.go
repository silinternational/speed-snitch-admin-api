package db

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/fillup/semver"
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"log"
	"net/http"
	"os"
)

const CASCADE = "CASCADE"
const NOACTION = "NO ACTION"
const SETNULL = "SET NULL"
const RESTRICT = "RESTRICT"

var Db *gorm.DB

var DatabaseTables = []interface{}{
	&domain.Contact{}, &domain.Country{}, &domain.Tag{}, &domain.Task{}, &domain.SpeedTestNetServer{},
	&domain.UserTags{}, &domain.User{}, &domain.Version{}, &domain.TaskLogSpeedTest{},
	&domain.TaskLogPingTest{}, &domain.TaskLogError{}, &domain.TaskLogRestart{}, &domain.TaskLogNetworkDowntime{},
	&domain.ReportingSnapshot{}, &domain.NamedServer{}, &domain.Node{}}

func GetDb() (*gorm.DB, error) {
	if Db == nil {
		host := domain.GetEnv("MYSQL_HOST", "localhost")
		user := domain.GetEnv("MYSQL_USER", "user")
		pass := domain.GetEnv("MYSQL_PASS", "pass")
		dbName := domain.GetEnv("MYSQL_DB", "test")
		dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, dbName)
		gdb, err := gorm.Open("mysql", dsn)
		if err != nil {
			return &gorm.DB{}, err
		}
		Db = gdb
		Db.SingularTable(true)
		Db.LogMode(false)
		Db.SetLogger(log.New(os.Stdout, "\r\n", 0))

	}
	return Db, nil
}

func AutoMigrateTables() error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	gdb.SingularTable(true)

	for _, table := range DatabaseTables {
		gdb.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(table)
		if gdb.Error != nil {
			return gdb.Error
		}
	}

	return CreateForeignKeys()
}

func CreateForeignKeys() error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	gdb.SingularTable(true)

	keys := []domain.ForeignKey{
		{
			ChildModel:  &domain.Contact{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.Node{},
			ChildField:  "running_version_id",
			ParentTable: "version",
			ParentField: "id",
			OnDelete:    SETNULL,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.Node{},
			ChildField:  "configured_version_id",
			ParentTable: "version",
			ParentField: "id",
			OnDelete:    SETNULL,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.Task{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.Task{},
			ChildField:  "named_server_id",
			ParentTable: "named_server",
			ParentField: "id",
			OnDelete:    RESTRICT,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.NamedServer{},
			ChildField:  "speed_test_net_server_id",
			ParentTable: "speed_test_net_server",
			ParentField: "id",
			OnDelete:    SETNULL,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogSpeedTest{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogSpeedTest{},
			ChildField:  "node_running_version_id",
			ParentTable: "version",
			ParentField: "id",
			OnDelete:    SETNULL,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogPingTest{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogPingTest{},
			ChildField:  "node_running_version_id",
			ParentTable: "version",
			ParentField: "id",
			OnDelete:    SETNULL,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogError{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogError{},
			ChildField:  "node_running_version_id",
			ParentTable: "version",
			ParentField: "id",
			OnDelete:    SETNULL,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogRestart{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.TaskLogNetworkDowntime{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.ReportingSnapshot{},
			ChildField:  "node_id",
			ParentTable: "node",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
		{
			ChildModel:  &domain.UserTags{},
			ChildField:  "user_id",
			ParentTable: "user",
			ParentField: "id",
			OnDelete:    CASCADE,
			OnUpdate:    NOACTION,
		},
	}

	for _, key := range keys {
		parentRef := fmt.Sprintf("%s(%s)", key.ParentTable, key.ParentField)
		gdb.Model(key.ChildModel).AddForeignKey(key.ChildField, parentRef, key.OnDelete, key.OnUpdate)
		if gdb.Error != nil {
			return gdb.Error
		}
	}

	return nil
}

func DropTables() error {
	db, err := GetDb()
	if err != nil {
		return err
	}

	db.SingularTable(true)
	db.Exec("SET FOREIGN_KEY_CHECKS=0")

	for _, table := range DatabaseTables {
		db.DropTable(table)
		if db.Error != nil {
			return db.Error
		}
	}

	// Need to manually drop many2many tables since they don't have their own models
	db.DropTable("node_tags")
	db.DropTable("user_tags")
	db.Exec("SET FOREIGN_KEY_CHECKS=1")
	return nil
}

func GetItem(itemObj interface{}, id uint) error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	notFound := gdb.Set("gorm:auto_preload", true).Unscoped().First(itemObj, id).RecordNotFound()
	if notFound {
		return gorm.ErrRecordNotFound
	}

	return gdb.Error
}

func GetNodeByMacAddr(macAddr string) (domain.Node, error) {
	node := domain.Node{
		MacAddr: macAddr,
	}

	err := FindOne(&node)
	if err != nil {
		return domain.Node{}, err
	}

	return node, nil
}

func GetSpeedTestNetServerByServerID(serverID string) (domain.SpeedTestNetServer, error) {
	server := domain.SpeedTestNetServer{
		ServerID: serverID,
	}

	err := FindOne(&server)
	if err != nil {
		return domain.SpeedTestNetServer{}, err
	}

	return server, nil
}

func GetCountryByCode(countryCode string) (domain.Country, error) {
	country := domain.Country{
		Code: countryCode,
	}

	err := FindOne(&country)
	if err != nil {
		return domain.Country{}, err
	}

	return country, nil
}

func ListItems(itemObj interface{}, order string) error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	gdb.Set("gorm:auto_preload", true).Unscoped().Order(order).Find(itemObj)

	return gdb.Error
}

func PutItem(itemObj interface{}) error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	newGdb := gdb.Save(itemObj)
	if newGdb.RecordNotFound() {
		return gorm.ErrRecordNotFound
	}

	errs := newGdb.GetErrors()
	if len(errs) > 0 {
		fmt.Fprintf(os.Stdout, "errors: %+v", errs)
		return errs[0]
	}

	return nil
}

func PutItemWithAssociations(itemObj interface{}, replacements []domain.AssociationReplacement) error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	tx := gdb.Begin()

	newGdb := tx.Save(itemObj)
	if newGdb.RecordNotFound() {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	errs := newGdb.GetErrors()
	if len(errs) > 0 {
		tx.Rollback()
		fmt.Fprintf(os.Stdout, "errors: %+v\n", errs)
		return errs[0]
	}

	for _, replace := range replacements {
		tx.Model(itemObj).Association(replace.AssociationName).Replace(replace.Replacement)
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
	}

	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return tx.Commit().Error
}

func DeleteItem(itemObj interface{}, id uint) error {
	err := GetItem(itemObj, id)
	if err != nil {
		return err
	}

	gdb, err := GetDb()
	if err != nil {
		return err
	}

	notFound := gdb.Unscoped().Delete(itemObj).RecordNotFound()
	if notFound {
		return gorm.ErrRecordNotFound
	}

	return gdb.Error
}

func FindOne(itemObj interface{}) error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	notFound := gdb.Set("gorm:auto_preload", true).Where(itemObj).First(itemObj).RecordNotFound()
	if notFound {
		return gorm.ErrRecordNotFound
	}

	return gdb.Error
}

func GetTaskLogForRange(itemObj interface{}, nodeId uint, rangeStart, rangeEnd int64) error {
	gdb, err := GetDb()
	if err != nil {
		return err
	}

	order := fmt.Sprintf("timestamp asc")
	where := fmt.Sprintf("node_id = ? AND timestamp between ? AND ?")
	gdb.Set("gorm:auto_preload", true).Order(order).Where(where, nodeId, rangeStart, rangeEnd).Find(itemObj)

	return gdb.Error
}

func GetUserFromRequest(req events.APIGatewayProxyRequest) (domain.User, error) {
	uuid, ok := req.Headers[domain.UserReqHeaderUUID]
	if !ok {
		return domain.User{}, fmt.Errorf("missing Header: %s", domain.UserReqHeaderUUID)
	}

	email, ok := req.Headers[domain.UserReqHeaderEmail]
	if !ok {
		return domain.User{}, fmt.Errorf("missing Header: %s", domain.UserReqHeaderEmail)
	}

	user := domain.User{
		Email: email,
	}

	err := FindOne(&user)
	if err != nil {
		return domain.User{}, err
	}

	// If first login, uuid will be empty and we need to set it
	if user.UUID == "" {
		user.UUID = uuid
		err := PutItem(&user)
		if err != nil {
			return user, err
		}
	} else if user.UUID != uuid {
		return domain.User{}, fmt.Errorf("user with email address %s exists, but UUID does not match. Received %s", user.Email, uuid)
	}

	return user, nil
}

func ListNamedServersByType(serverType string) ([]domain.NamedServer, error) {

	gdb, err := GetDb()
	if err != nil {
		return []domain.NamedServer{}, err
	}

	order := fmt.Sprintf("name asc")
	serverList := []domain.NamedServer{}
	where := fmt.Sprintf("server_type = ?")
	gdb.Set("gorm:auto_preload", true).Unscoped().Order(order).Where(where, serverType).Find(&serverList)
	return serverList, gdb.Error
}

// AreTagsValid returns a bool based on this ...
//  - if the input is empty, then true
//  - if there is an error getting the tags from the database, then false
//  - if any of the tags do not have a UID that matches one that's in the Db, then false
//  - if all of the tags have a UID that matches one that's in the Db, then true
func AreTagsValid(tags []domain.Tag) bool {
	if len(tags) == 0 {
		return true
	}

	ids := []uint{}
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}

	db, err := GetDb()
	if err != nil {
		return false
	}

	var foundTags []domain.Tag
	db.Where("id in (?)", ids).Find(&foundTags)
	if db.Error != nil {
		return false
	}

	return len(tags) == len(foundTags)
}

// GetLatestVersion iterates through version in Db to return only the latest version
func GetLatestVersion() (domain.Version, error) {
	var versions []domain.Version
	err := ListItems(&versions, "number asc")
	if err != nil {
		return domain.Version{}, err
	}

	var latest domain.Version

	for _, version := range versions {
		if latest.Number == "" {
			latest = version
			continue
		}

		isNewer, err := semver.IsNewer(latest.Number, version.Number)
		if err != nil {
			return domain.Version{}, err
		}
		if isNewer {
			latest = version
		}
	}

	return latest, nil
}

// GetAuthorizationStatus returns 0, nil for users that are authorized to use the object
func GetAuthorizationStatus(req events.APIGatewayProxyRequest, permissionType string, objectTags []domain.Tag) (int, string) {
	user, err := GetUserFromRequest(req)
	if err != nil {
		return http.StatusBadRequest, err.Error()
	}

	if user.Role == domain.PermissionSuperAdmin {
		return 0, ""
	}

	if permissionType == domain.PermissionSuperAdmin {

		fmt.Fprintf(
			os.Stdout,
			"Attempt at unauthorized access at path: %s.\n  User: %s.\n  User is not a superAdmin.\n",
			req.Path,
			user.Email,
		)
		return http.StatusForbidden, http.StatusText(http.StatusForbidden)
	}

	if permissionType == domain.PermissionTagBased {
		tagsOverlap := domain.DoTagsOverlap(user.Tags, objectTags)
		if tagsOverlap {
			return 0, ""
		}

		fmt.Fprintf(
			os.Stdout,
			"Attempt at unauthorized access at path: %s.\n  User: %s.\n  User Tags: %v.\n  Object Tags: %v.\n",
			req.Path,
			user.Email,
			user.Tags,
			objectTags,
		)

		return http.StatusForbidden, http.StatusText(http.StatusForbidden)
	}

	return http.StatusInternalServerError, "Invalid permission type requested: " + permissionType
}

func GetSnapshotsForRange(interval string, nodeId uint, rangeStart, rangeEnd int64) ([]domain.ReportingSnapshot, error) {
	gdb, err := GetDb()
	if err != nil {
		return []domain.ReportingSnapshot{}, err
	}

	var snapshots []domain.ReportingSnapshot
	where := "`node_id` = ? AND `interval` = ? AND `timestamp` between ? AND ?"
	gdb.Set("gorm:auto_preload", true).Order("timestamp asc").Where(where, nodeId, interval, rangeStart, rangeEnd).Find(&snapshots)

	return snapshots, gdb.Error
}

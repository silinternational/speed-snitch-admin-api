package testutils

import (
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"testing"
)

var SuperAdmin = domain.User{
	Role:  domain.UserRoleSuperAdmin,
	Email: "super@admin.com",
	UUID:  "11111111-1111-1111-1111-111111111111",
	Name:  "Super Admin",
	Model: gorm.Model{
		ID: 1,
	},
}

var AdminUser = domain.User{
	Role:  domain.UserRoleAdmin,
	Email: "normal@admin.com",
	UUID:  "22222222-2222-2222-2222-222222222222",
	Name:  "Normal Admin",
	Model: gorm.Model{
		ID: 2,
	},
}

func MigrateTables(t *testing.T) {
	err := db.AutoMigrateTables()
	if err != nil {
		t.Error("Error migrating tables: ", err.Error())
	}
}

func DropTables(t *testing.T) {
	err := db.DropTables()
	if err != nil {
		t.Error("Error dropping tables: ", err.Error())
	}
}

func ResetDb(t *testing.T) {
	DropTables(t)
	MigrateTables(t)
	CreateSuperAdmin(t)
}

func CreateSuperAdmin(t *testing.T) {
	err := db.PutItem(&SuperAdmin)
	if err != nil {
		t.Error("Error creating super admin: ", err.Error())
	}
}

func CreateAdminUser(t *testing.T) {
	err := db.PutItem(&AdminUser)
	if err != nil {
		t.Error("Error creating admin user: ", err.Error())
	}
}

func GetSuperAdminReqHeader() map[string]string {
	return map[string]string{
		"x-user-uuid": SuperAdmin.UUID,
		"x-user-mail": SuperAdmin.Email,
	}
}

func GetAdminUserReqHeader() map[string]string {
	return map[string]string{
		"x-user-uuid": AdminUser.UUID,
		"x-user-mail": AdminUser.Email,
	}
}

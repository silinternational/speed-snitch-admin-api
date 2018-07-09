package db

import (
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"testing"
)

func migrateTables(t *testing.T) {
	err := AutoMigrateTables()
	if err != nil {
		t.Error("Error migrating tables: ", err.Error())
	}
}

func dropTables(t *testing.T) {
	err := DropTables()
	if err != nil {
		t.Error("Error dropping tables: ", err.Error())
	}
}

func resetDb(t *testing.T) {
	dropTables(t)
	migrateTables(t)
}

func TestAreTagsValid(t *testing.T) {
	resetDb(t)
	validTags := []domain.Tag{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "one",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Name: "two",
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Name: "three",
		},
	}

	invalidTags := []domain.Tag{
		{
			Model: gorm.Model{
				ID: 4,
			},
			Name: "four",
		},
		{
			Model: gorm.Model{
				ID: 5,
			},
			Name: "five",
		},
	}

	for _, tag := range validTags {
		PutItem(&tag)
	}

	shouldBeValid := []domain.Tag{
		validTags[0],
		validTags[1],
	}

	if !AreTagsValid(shouldBeValid) {
		t.Error("Valid tags failed check if valid")
	}

	shouldNotBeValid := []domain.Tag{
		validTags[1],
		invalidTags[1],
	}

	if AreTagsValid(shouldNotBeValid) {
		t.Error("Invalid tags passed as valid")
	}
}

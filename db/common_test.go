package db

import (
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"testing"
)

func TestAreTagsValid(t *testing.T) {
	DropTables()
	AutoMigrateTables()

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

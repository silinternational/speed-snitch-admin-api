package db

import (
	"github.com/jinzhu/gorm"
	"github.com/silinternational/speed-snitch-admin-api"
	"testing"
	"time"
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

func TestDeleteItem(t *testing.T) {

	DropTables()
	AutoMigrateTables()

	namedServer := domain.NamedServer{
		Name:       "Test Server",
		ServerType: domain.ServerTypePing,
		ServerHost: "test.host.org",
	}

	err := PutItem(&namedServer)
	if err != nil {
		t.Errorf("Error saving fixture. %s", err.Error())
		return
	}

	var allNamedServers []domain.NamedServer
	err = ListItems(&allNamedServers, "")
	if err != nil {
		t.Errorf("Error trying to check fixture. %s", err.Error())
		return
	}

	if len(allNamedServers) != 1 {
		t.Errorf("Wrong number of fixtures loaded.  Expected 1, but got %d.", len(allNamedServers))
		return
	}

	// Here's the real test
	err = DeleteItem(&namedServer, namedServer.ID)
	if err != nil {
		t.Errorf("Error deleting the namedserver. %s", err.Error())
		return
	}

	err = ListItems(&allNamedServers, "")
	if err != nil {
		t.Errorf("Error trying to check fixture. %s", err.Error())
		return
	}

	if len(allNamedServers) != 0 {
		t.Errorf("Wrong number of namedservers remaining.  Expected 0, but got %d.", len(allNamedServers))
		return
	}
}

func TestListMIANodes(t *testing.T) {
	DropTables()
	AutoMigrateTables()

	daysMissing := 2
	tNow := time.Now().UTC()

	miaNode1 := domain.Node{
		MacAddr:  "11:22:33:44:55:ba",
		LastSeen: tNow.AddDate(0, 0, -(daysMissing + 1)).Format(time.RFC3339),
	}

	miaNode2 := domain.Node{
		MacAddr:  "11:22:33:44:55:bb",
		LastSeen: tNow.AddDate(0, 0, -(daysMissing + 2)).Format(time.RFC3339),
	}

	goodNode := domain.Node{
		MacAddr:  "11:22:33:44:55:aa",
		LastSeen: tNow.AddDate(0, 0, -(daysMissing - 1)).Format(time.RFC3339),
	}

	for _, nodePtr := range []*domain.Node{&miaNode1, &miaNode2, &goodNode} {
		err := PutItem(nodePtr)
		if err != nil {
			t.Errorf("Error saving fixture. %s", err.Error())
			return
		}
	}

	results, err := ListMIANodes(daysMissing)
	if err != nil {
		t.Errorf("Got an unexpected error.\n %s", err.Error())
		return
	}

	expected := []domain.Node{miaNode1, miaNode2}
	if len(results) != 2 || results[0].ID != miaNode1.ID || results[1].ID != miaNode2.ID {
		t.Errorf("Did not get expected results. \nExpected: %+v\n But got: %+v", expected, results)
	}
}

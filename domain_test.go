package domain

import (
	"testing"
)

type testMACAddr struct {
	MA1         string
	MA2         string
	ExpectError bool
}

func TestCleanMACAddress(t *testing.T) {

	testMAs := []testMACAddr{
		{
			MA1:         "112233AAbbCC",
			MA2:         "112233aabbcc",
			ExpectError: false,
		},
		{
			MA1:         "11:22:33:AA:bb:CC",
			MA2:         "11:22:33:aa:bb:cc",
			ExpectError: false,
		},
		{
			MA1:         "11:22-33:AA-bb:CC",
			MA2:         "11:22-33:aa-bb:cc",
			ExpectError: false,
		},
		{
			MA1:         "112233AAbb",
			MA2:         "",
			ExpectError: true, // too short
		},
		{
			MA1:         "112233AAbbCCDD",
			MA2:         "",
			ExpectError: true, // too long
		},
		{
			MA1:         "G12233AAbbCC",
			MA2:         "",
			ExpectError: true, // bad letter
		},
		{
			MA1:         "11:22-33aabbcc",
			MA2:         "",
			ExpectError: true, // not enough delimiters
		},
	}

	for _, ma := range testMAs {

		resultMA, err := CleanMACAddress(ma.MA1)
		if ma.ExpectError && err == nil {
			t.Errorf("For test MAC Address %s expected an error but did not get one.", ma.MA1)
		} else if !ma.ExpectError && err != nil {
			t.Errorf("For test MAC Address %s did not expect an error but got one:\n\t%s.", ma.MA1, err.Error())
		}

		if resultMA != ma.MA2 {
			t.Errorf(`For test MAC Address %s expected: "%s", but got "%s".`, ma.MA1, ma.MA2, resultMA)
		}
	}
}

func TestMergeStructsFromJson(t *testing.T) {
	versionNumber := "0.0.1"
	newDescription := "New Version"

	oldObj := Version{
		versionNumber,
		"Old Version",
	}

	newJson := `{
	"Description": "` + newDescription + `"
}`

	err := MergeStructsFromJson(oldObj, newJson)
	if err != nil {
		t.Errorf("Did not expect an error but got \n\t%s", err.Error())
		return
	}

	results := oldObj.Number

	if versionNumber != results {
		t.Errorf("Bad Number. Expected %s, but got %s", versionNumber, results)
		return
	}

	results = oldObj.Description
	if newDescription != results {
		t.Errorf("Bad Description. Expected %s, but got %s", newDescription, results)
		return
	}
}

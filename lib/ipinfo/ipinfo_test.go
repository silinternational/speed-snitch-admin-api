package ipinfo

import "testing"

func TestGetIPInfo(t *testing.T) {
	expected := IPDetails{
		IP:       "8.8.8.8",
		Hostname: "google-public-dns-a.google.com",
		Loc:      "37.385999999999996,-122.0838",
		Org:      "AS15169 Google Inc.",
		City:     "Mountain View",
		Region:   "California",
		Country:  "US",
		Phone:    "650",
	}

	ipDetails, err := GetIPInfo("8.8.8.8")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if ipDetails.City != expected.City {
		t.Error(ipDetails)
		t.Fail()
	}
}

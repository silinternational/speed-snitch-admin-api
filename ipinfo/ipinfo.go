package ipinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const ApiURL = "http://ipinfo.io"

type IPDetails struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Phone    string `json:"phone"`
}

// GetIPInfo calls ipinfo.io API to return geo location information for given IP address
func GetIPInfo(ipAddress string) (IPDetails, error) {
	url := fmt.Sprintf("%s/%s", ApiURL, ipAddress)
	headers := map[string]string{
		"Accept": "application/json",
	}
	resp, err := CallAPI(http.MethodGet, url, "", headers)
	defer resp.Body.Close()
	if err != nil {
		return IPDetails{}, err
	}

	var details IPDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return IPDetails{}, err
	}

	return details, nil
}

// CallAPI creates a http.Request object, attaches headers to it and makes the
// requested api call.
func CallAPI(method, url, postData string, headers map[string]string) (*http.Response, error) {
	var err error
	var req *http.Request

	if postData != "" {
		req, err = http.NewRequest(method, url, strings.NewReader(postData))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	} else if resp.StatusCode >= 300 {
		return resp, fmt.Errorf("API returned an error. \n\tMethod: %s, \n\tURL: %s, \n\tCode: %v, \n\tStatus: %s \n\tBody: %s",
			method, url, resp.StatusCode, resp.Status, postData)
	}

	return resp, nil
}

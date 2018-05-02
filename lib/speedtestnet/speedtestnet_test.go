package speedtestnet

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSTNetServers(t *testing.T) {

	mux := http.NewServeMux()

	testServer := httptest.NewServer(mux)

	respBody := `<?xml version="1.0" encoding="UTF-8"?>
<settings>
<servers><speedtestnetserver url="http://88.84.191.230/speedtest/upload.php" lat="70.0733" lon="29.7497" name="Vadso" country="Norway" cc="NO" sponsor="Varanger KraftUtvikling AS" id="4600"  url2="http://speedmonster.varangerbynett.no/speedtest/upload.php" host="88.84.191.230:8080" />
<speedtestnetserver url="http://speedtest.nornett.net/speedtest/upload.php" lat="69.9403" lon="23.3106" name="Alta" country="Norway" cc="NO" sponsor="Nornett AS" id="4961"  url2="http://speedtest2.nornett.net/speedtest/upload.php" host="speedtest.nornett.net:8080" />
<speedtestnetserver url="http://speedo.eltele.no/speedtest/upload.php" lat="69.9403" lon="23.3106" name="Alta" country="Norway" cc="NO" sponsor="Eltele AS" id="3433"  host="speedo.eltele.no:8080" />
<speedtestnetserver url="http://tos.speedtest.as2116.net/speedtest/upload.php" lat="69.6492" lon="18.9553" name="Tromsø" country="Norway" cc="NO" sponsor="Broadnet" id="11786"  host="tos.speedtest.as2116.net:8080" />
</servers>
</settings>`

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-type", "test/xml")
		w.WriteHeader(200)
		fmt.Fprintf(w, respBody)
	})

	servers, err := GetSTNetServers(testServer.URL)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	expectedLen := 4
	resultsLen := len(servers)
	if resultsLen != 4 {
		t.Errorf("Wrong number of servers. Expected: %d. Got: %d", expectedLen, resultsLen)
		t.Fail()
	}

	expectedIDs := []string{"4600", "4961", "3433", "11786"}
	for index, nextServer := range servers {
		result := nextServer.ServerID
		expected := expectedIDs[index]
		if result != expected {
			t.Errorf("Wrong speedtestnetserver ID at index %d. Expected: %s. Got: %s", index, expected, result)
			t.Fail()
			break
		}
	}
}

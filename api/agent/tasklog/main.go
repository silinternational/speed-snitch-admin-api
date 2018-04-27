package tasklog

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
)

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var taskLogEntries []domain.TaskLogEntry
	err := json.Unmarshal([]byte(req.Body), &taskLogEntries)
	if err != nil {
		return domain.ClientError(http.StatusUnprocessableEntity, err.Error())
	}

	for _, entry := range taskLogEntries {
		var speedTestServer domain.SpeedTestNetServer
		err := db.GetItem(domain.SpeedTestNetServerTable, "ID", entry.ServerID, &speedTestServer)
		if err != nil {
			// How to handle? Assume log somewhere but proceed processing taskLogEntries
		} else {
			entry.ServerCountry = speedTestServer.Country
			entry.ServerCoordinates = fmt.Sprintf("%s,%s", speedTestServer.Lat, speedTestServer.Lon)
			tableAlias := ""

			if entry.Latency != 0 {
				tableAlias = domain.TaskLogLatencyTable
			} else if entry.ErrorCode != "" {
				//tableAlias = domain.TaskLogErrorTable
			} else if entry.Upload != 0 {
				tableAlias = domain.TaskLogSpeedTable
			} else {
				// handle as error?
			}

			db.PutItem(tableAlias, entry)
		}
	}

	// Return a response with a 200 OK status and the JSON book record
	// as the body.
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Body:       "",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

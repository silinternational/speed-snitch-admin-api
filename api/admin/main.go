package admin

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/silinternational/speed-snitch-admin-api"
	"github.com/silinternational/speed-snitch-admin-api/db"
	"net/http"
	"strings"
)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	pathParts := strings.Split(req.Path, "/")
	subPath := pathParts[1]

	switch subPath {
	case "namedserver":
		return namedserverRouter(req)
	case "node":
		return nodeRouter(req)
	case "report":
		return reportRouter(req)
	case "reportingevent":
		return reportingeventRouter(req)
	case "speedtestnetserver":
		return speedtestnetserverRouter(req)
	case "tag":
		return tagRouter(req)
	case "user":
		return userRouter(req)
	case "version":
		return versionRouter(req)

	default:
		return domain.ClientError(http.StatusNotFound, "Bad path: "+req.Path)
	}
}

func main() {
	defer db.Db.Close()
	lambda.Start(router)
}

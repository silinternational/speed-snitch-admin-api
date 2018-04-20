build:
	docker-compose run go go build -ldflags="-s -w" -o bin/hello              api/agent/hello/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/tag                api/admin/tag/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/node               api/admin/node/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/speedtestnetserver api/admin/speedtestnetserver/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/user               api/admin/user/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/version            api/admin/version/main.go

slsdeploy: slsdeployadmin slsdeployagent

slsdeployadmin:
	docker-compose run sls bash -c "cd api/admin && sls deploy"

slsdeployagent:
	docker-compose run sls bash -c "cd api/agent && sls deploy"

deploy: build slsdeploy

dep:
	docker-compose run go dep ensure

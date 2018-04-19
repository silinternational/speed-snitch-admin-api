build:
	docker-compose run go go build -ldflags="-s -w" -o bin/hello              api/hello/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/tag                api/tag/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/node               api/node/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/speedtestnetserver api/speedtestnetserver/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/user               api/user/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/version            api/version/main.go

slsdeploy:
	docker-compose run sls sls deploy

deploy: build
	make slsdeploy

dep:
	docker-compose run go dep ensure

build:
	docker-compose run go go build -ldflags="-s -w" -o bin/hello hello/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/tag tag/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/node node/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/user user/main.go
	docker-compose run go go build -ldflags="-s -w" -o bin/version user/main.go

slsdeploy:
	docker-compose run sls sls deploy

deploy: build
	make slsdeploy

dep:
	docker-compose run go dep ensure
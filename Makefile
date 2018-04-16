build:
	docker-compose run go go build -ldflags="-s -w" -o bin/hello hello/main.go

deploy: build
	docker-compose run sls sls deploy

dep:
	docker-compose run go dep ensure
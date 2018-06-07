build:
	docker-compose run app ./codeship/build.sh

slsdeploy: slsdeployadmin slsdeployagent

slsdeployadmin:
	docker-compose run app bash -c "cd api/admin && sls deploy"

slsdeployagent:
	docker-compose run app bash -c "cd api/agent && sls deploy"

deploy: build slsdeploy

dep:
	docker-compose run app dep ensure

test:
	docker-compose run test ./codeship/test.sh

codeshipsetup: dep build

dynamo:
	docker-compose up -d dynamo
	sleep 5
	docker-compose run init-dynamo
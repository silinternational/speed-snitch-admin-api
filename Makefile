build:
	docker-compose run app ./codeship/build.sh

slsdeploy: slsdeployadmin slsdeployagent

slsdeployadmin:
	docker-compose run app bash -c "cd api/admin && sls deploy"

slsdeployagent:
	docker-compose run app bash -c "cd api/agent && sls deploy"

deploy: build slsdeploy

deployagent: buildagent slsdeployagent

deployadmin: buildadmin slsdeployadmin

dep:
	docker-compose run app dep ensure

test:
	docker-compose run app ./codeship/test.sh

codeshipsetup: dep build
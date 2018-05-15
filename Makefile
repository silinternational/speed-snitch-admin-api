build:
	docker-compose run go ./codeship/build.sh

slsdeploy: slsdeployadmin slsdeployagent

slsdeployadmin:
	docker-compose run sls bash -c "cd api/admin && sls deploy"

slsdeployagent:
	docker-compose run sls bash -c "cd api/agent && sls deploy"

deploy: build slsdeploy

deployagent: buildagent slsdeployagent

deployadmin: buildadmin slsdeployadmin

dep:
	docker-compose run go dep ensure

test:
	docker-compose run go ./codeship/test.sh

codeshipsetup: dep build
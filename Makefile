build:
	docker-compose run --rm app ./codeship/build.sh

deploy: build
	docker-compose run --rm app ./codeship/deploy-dev.sh

test:
	docker-compose run --rm test ./codeship/test.sh

clean:
	docker-compose kill
	docker-compose rm -f

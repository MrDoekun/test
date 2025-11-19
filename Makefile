build:
	./docker/build.sh

.PHONY: vendor
vendor:
	docker run --rm -e GO111MODULE=on -v "$(PWD)":/go/src/amartha golang-dev sh -c 'go mod tidy && go mod vendor'

up:
	@make build
	@make vendor
	docker-compose -f docker-compose.yml up -d

down:
	docker-compose down
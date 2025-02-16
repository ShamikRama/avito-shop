IMAGE_NAME = avitoservice
LINTER=golangci-lint

default: up

build:
	docker build -t $(IMAGE_NAME) .

up: build
	docker-compose up

clean:
	docker-compose down

test:
	sh test.sh

lint:
	$(LINTER) run --config .golangci.yaml


.PHONY: default build up

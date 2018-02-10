.PHONY: $(shell ls -d *)

default:
	@echo "Usage: make [command]"

build:
	[ "${TAG}" != "" ] || (echo "USAGE: make build TAG=TAGID"; false)
	docker build -t kavehmz/jobber:${TAG} .
	docker push kavehmz/jobber:${TAG}

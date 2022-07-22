.PHONY: clean build test shell release

all: clean build

CONTAINER_NAME=jlentink/frontend-deploy
CONTAINER_VERSION=1.0.0
MOUNT_VOLUME?=$(shell pwd)

release: clean
	goreleaser release --rm-dist

clean:
	-rm -rf dist
	-rm -f frontend-deploy

test:
	goreleaser --skip-publish --skip-validate --rm-dist --snapshot

build:
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o frontend-deploy .

shell:
	docker run -it --rm --entrypoint "/bin/sh" -v ${MOUNT_VOLUME}:/app ${CONTAINER_NAME}:${CONTAINER_VERSION}

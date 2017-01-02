# Just builds
DIR := ${CURDIR}

dep:
	glide install --strip-vendor

dep2:
	# glide messes up on some of these things while trying to strip-vendor, so deleting them
	rm -rf vendor/github.com/heroku/docker-registry-client/vendor
	rm -rf vendor/github.com/docker/docker/hack
	rm -rf vendor/github.com/docker/docker/project

cp-runner:
	# copies iron-io/runner into functions/vendor, good for dual development
	rsync -av --exclude='vendor' --exclude='examples' --exclude='docs' --exclude='.git' ../runner/ vendor/github.com/iron-io/runner

build:
	go build -o functions

build-docker:
	docker run --rm -v $(DIR):/go/src/github.com/iron-io/functions -w /go/src/github.com/iron-io/functions iron/go:dev go build -o functions-alpine
	docker build -t iron/functions:latest .

test:
	go test -v $(shell go list ./... | grep -v vendor | grep -v examples | grep -v tool | grep -v fn)
	cd fn && $(MAKE) test

test-datastore:
	cd api/datastore && go test -v

test-docker:
	docker run -ti --privileged --rm -e LOG_LEVEL=debug \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v $(DIR):/go/src/github.com/iron-io/functions \
	-w /go/src/github.com/iron-io/functions iron/go:dev go test \
	-v $(shell go list ./... | grep -v vendor | grep -v examples | grep -v tool | grep -v fn | grep -v datastore)

run:
	./functions

run-docker: build-docker
	docker run --rm --privileged -it -e LOG_LEVEL=debug -e "DB_URL=bolt:///app/data/bolt.db" -v $(DIR)/data:/app/data -p 8080:8080 iron/functions

all: dep build
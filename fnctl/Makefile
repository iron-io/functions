all: vendor
	GOOS=linux go build -o fnctl
	docker build -t iron/fnctl .
	docker push iron/fnctl

vendor:
	glide install --strip-vendor


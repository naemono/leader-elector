all: push

# Docker Image URL to use all building/pushing image targets
VERSION ?= latest
PREFIX ?= naemono
NAME := leader-elector
IMG := $(PREFIX)/$(NAME):$(VERSION)

server:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o server main.go

docker-build: server
	docker build --pull -t ${IMG} .

docker-push:
	docker push ${IMG}

clean:
	rm -f server

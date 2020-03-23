all: push

# Docker Image URL to use all building/pushing image targets
VERSION ?= latest
PREFIX ?= mmontg1
NAME := leader-elector
IMG := $(PREFIX)/$(NAME):$(VERSION)

server:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o server main.go

docker-build: server
	docker build --pull -t ${IMG} .

docker-push:
	docker push ${IMG}
	docker push $(PREFIX)/$(NAME):latest

clean:
	rm -f server

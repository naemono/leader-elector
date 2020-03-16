all: push

# 0.0 shouldn't clobber any released builds
# current latest is 0.5
TAG = 0.0
PREFIX = naemono/leader-elector

server:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o server main.go

container: server
	docker build --pull -t $(PREFIX):$(TAG) .

push: container
	docker push $(PREFIX):$(TAG)

clean:
	rm -f server

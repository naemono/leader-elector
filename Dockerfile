# Build the leader-elector binary
FROM golang:1.14 as builder

# Copy in the go src
WORKDIR /usr/local/src/leader-elector
COPY go.sum go.sum
COPY go.mod go.mod
COPY pkg pkg
COPY main.go main.go

# Build
RUN GOPROXY=https://proxy.golang.org CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -o /server

# Copy the server into a thin image
FROM alpine:3.11
ENV USER=leader-elector
WORKDIR /
COPY --from=builder /server .
COPY run.sh /run.sh
RUN addgroup -g 1000 ${USER} && \
  adduser -D -g "${USER} user" -h "/home/leader-elector" -G "${USER}" -u 1000 ${USER} && \
  chown -R ${USER}:${USER} /home/leader-elector /server /run.sh
ENTRYPOINT ["/run.sh"]

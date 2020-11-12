FROM golang:1.15.2-buster

WORKDIR /server

COPY go.mod go.sum ./

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon -log-prefix=false -include=*.yml -build="go build -o bin/main cmd/main.go" -command="bin/main --config config/config.yml"

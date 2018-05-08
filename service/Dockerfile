FROM golang:1.10

ADD . /go/src/github.com/Tahler/service-grapher/service

RUN go install github.com/Tahler/service-grapher/service

ENTRYPOINT "/go/bin/service"

EXPOSE 8080

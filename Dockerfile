FROM golang

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/Tahler/service-grapher

COPY . .
RUN dep ensure -vendor-only

RUN go install github.com/Tahler/service-grapher/service

ENTRYPOINT "/go/bin/service"

EXPOSE 8080

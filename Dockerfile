FROM golang:1.10.2 AS builder

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/Tahler/service-grapher

COPY Gopkg.toml Gopkg.toml
COPY Gopkg.lock Gopkg.lock
RUN dep ensure -vendor-only

COPY pkg pkg
COPY service service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./service

# TODO: Use minimal image once curl is no longer needed for debugging.
# FROM scratch
# TODO: Add this if HTTPS is needed.
# RUN apk --no-cache add ca-certificates

FROM ubuntu
RUN apt-get update
RUN apt-get install -y curl
WORKDIR /root
COPY --from=builder /go/src/github.com/Tahler/service-grapher/app .
EXPOSE 8080

CMD ["./app"]

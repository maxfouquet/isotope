FROM golang:1.10.3 AS builder

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/Tahler/isotope/automation/convert

COPY . .

RUN dep ensure -vendor-only

RUN CGO_ENABLED=0 GOOS=linux \
  go build -a -installsuffix cgo -o ./main ./main.go

FROM scratch
COPY --from=builder \
  /go/src/github.com/Tahler/isotope/automation/convert/main \
  /usr/local/bin/convert

ENTRYPOINT ["/usr/local/bin/convert"]

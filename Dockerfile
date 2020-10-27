FROM golang:1.15 AS builder
RUN mkdir /app
RUN mkdir /go/src/plenuslb

COPY . /go/src/plenuslb
WORKDIR /go/src/plenuslb

ENV GO111MODULE=on

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./pkg/cmd/controller/main.go

## ca-certificates
FROM alpine as certs
RUN apk update && apk add ca-certificates


FROM busybox:1.31
COPY --from=certs /etc/ssl/certs /etc/ssl/certs

RUN mkdir /app

WORKDIR /app

COPY --from=builder /app/main .
CMD ["./main"]

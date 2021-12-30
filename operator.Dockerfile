FROM golang:1.16 AS builder
RUN mkdir /app
RUN mkdir /go/src/plenuslb

COPY . /go/src/plenuslb
WORKDIR /go/src/plenuslb

ENV GO111MODULE=on

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./pkg/cmd/operator/main.go

FROM busybox:1.31

RUN mkdir /app

WORKDIR /app

COPY --from=builder /app/main .
CMD ["./main"]

FROM golang:1.22 as builder

RUN mkdir -p /go/src/github.com/vinit-chauhan/load-balancer
WORKDIR /go/src/github.com/vinit-chauhan/load-balancer

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o /go/bin/load-balancer


FROM debian:bookworm-slim

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY --from=builder /go/bin/load-balancer .

COPY config.yml .

RUN chmod +x ./load-balancer

CMD ["./load-balancer"]
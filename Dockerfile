FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a cgo -o kube-watcher ./cmd/main.go

FROM alpine:latest AS run

WORKDIR /app

COPY --from=builder /app/kube-watcher .

CMD ["./kube-watcher"]

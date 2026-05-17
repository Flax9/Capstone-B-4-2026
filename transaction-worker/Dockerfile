FROM golang:1.26-alpine AS builder
WORKDIR /app
ENV GOWORK=off
COPY proto ./proto/
COPY transaction-worker ./transaction-worker/
WORKDIR /app/transaction-worker
RUN go mod download
RUN go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/transaction-worker/server .
CMD ["./server"]

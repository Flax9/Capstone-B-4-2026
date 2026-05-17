FROM golang:1.26-alpine AS builder
WORKDIR /app
ENV GOWORK=off
COPY proto ./proto/
COPY transaction-service ./transaction-service/
WORKDIR /app/transaction-service
RUN go mod download
RUN go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/transaction-service/server .
EXPOSE 9003
CMD ["./server"]

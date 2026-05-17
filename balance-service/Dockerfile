FROM golang:1.26-alpine AS builder
WORKDIR /app
ENV GOWORK=off
COPY proto ./proto/
COPY balance-service ./balance-service/
WORKDIR /app/balance-service
RUN go mod download
RUN go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/balance-service/server .
EXPOSE 9002
CMD ["./server"]

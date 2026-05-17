FROM golang:1.26-alpine AS builder
WORKDIR /app
ENV GOWORK=off
COPY proto ./proto/
COPY audit-worker ./audit-worker/
WORKDIR /app/audit-worker
RUN go mod download
RUN go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/audit-worker/server .
CMD ["./server"]

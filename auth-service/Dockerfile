FROM golang:1.26-alpine AS builder
WORKDIR /app
ENV GOWORK=off
COPY proto ./proto/
COPY shared ./shared/
COPY auth-service ./auth-service/
WORKDIR /app/auth-service
RUN go mod download
RUN go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auth-service/server .
EXPOSE 9001
CMD ["./server"]

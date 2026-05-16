FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
# Download semua deps termasuk indirect
RUN go mod download all
COPY . .
# Gunakan GONOSUMCHECK agar tidak membutuhkan go.sum yang sempurna
ENV GONOSUMCHECK=*
RUN go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 9003
CMD ["./server"]

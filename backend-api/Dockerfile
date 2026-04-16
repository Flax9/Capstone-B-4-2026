FROM golang:1.26-alpine AS builder

WORKDIR /app

# Tarik Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Salin Kodingan dan Kompilasi
COPY . .
RUN go build -o server main.go

# Stage Final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .

# API Fiber Golang kita berada di 9000
EXPOSE 9000

CMD ["./server"]
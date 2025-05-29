FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/Darwin3-device-plugin cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/Darwin3-device-plugin .

ENTRYPOINT [ "./Darwin3-device-plugin" ]
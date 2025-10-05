FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/gateway ./src/gateway/cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/gateway /app/gateway

COPY src/gateway/configs /app/configs

EXPOSE 8000

ENTRYPOINT ["/app/gateway", "--config", "/app/configs"]

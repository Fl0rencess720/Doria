FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/memory ./src/services/memory/cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/memory /app/memory

COPY src/services/memory/configs /app/configs

EXPOSE 9005

ENTRYPOINT ["/app/memory", "--config", "/app/configs"]
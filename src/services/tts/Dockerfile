FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/tts ./src/services/tts/cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/tts /app/tts

COPY src/services/tts/configs /app/configs

EXPOSE 9003

ENTRYPOINT ["/app/tts", "--config", "/app/configs"]
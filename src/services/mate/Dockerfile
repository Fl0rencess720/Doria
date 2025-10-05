FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/mate ./src/services/mate/cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/mate /app/mate

COPY src/services/mate/configs /app/configs

EXPOSE 9004

ENTRYPOINT ["/app/mate", "--config", "/app/configs"]
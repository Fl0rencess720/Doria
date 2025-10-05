FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/user ./src/services/user/cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/user /app/user

COPY src/services/user/configs /app/configs

EXPOSE 9002

ENTRYPOINT ["/app/user", "--config", "/app/configs"]
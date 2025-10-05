FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/image ./src/services/image/cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/image /app/image

COPY src/services/image/configs /app/configs

EXPOSE 9001

ENTRYPOINT ["/app/image", "--config", "/app/configs"]
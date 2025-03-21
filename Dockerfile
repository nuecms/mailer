FROM golang:1.18-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mailer .

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl tzdata

WORKDIR /app
COPY --from=builder /build/mailer .
COPY config.example.json /app/config.example.json

# 创建邮件目录
RUN mkdir -p /app/emails/failed

EXPOSE 25 8025
VOLUME ["/app/emails", "/app/config.json"]

CMD ["/app/mailer", "-config", "/app/config.json"]

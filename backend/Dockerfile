# --- build stage ---
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /linkvault-backend .

# --- runtime stage ---
FROM alpine:latest

# fetcherパッケージがユーザー入力URLへhttpsアクセスするため、
# TLS検証用のCA証明書が必要（無いと外部ページのメタデータ取得が
# コンテナ上でだけ静かに失敗する）。
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder /linkvault-backend .

USER appuser

EXPOSE 8080

ENTRYPOINT ["./linkvault-backend"]

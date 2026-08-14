# ---------------------------------------------------------------- build
FROM golang:1.26-alpine AS builder

WORKDIR /src

# แยก layer ของ dependency ไว้ก่อน เพื่อให้ cache ทำงานเมื่อแก้แค่โค้ด
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 ทำให้ได้ binary แบบ static รันบน image ที่ไม่มี libc ได้
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/api  ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

# ---------------------------------------------------------------- runtime
FROM alpine:3.20

# ca-certificates สำหรับเรียก HTTPS, tzdata เพื่อให้เวลาไทยถูกต้อง
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 appuser

ENV TZ=Asia/Bangkok

WORKDIR /app
COPY --from=builder /out/api  /app/api
COPY --from=builder /out/seed /app/seed

# โฟลเดอร์เก็บสลิป/รูปภาพ ควร mount volume ทับใน production
RUN mkdir -p /app/uploads && chown -R appuser:appuser /app

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/api"]

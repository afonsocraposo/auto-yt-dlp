FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auto-yt-dlp .

# Final stage
FROM alpine:latest

RUN apk add --no-cache \
    python3 \
    py3-pip \
    ffmpeg \
    ca-certificates \
    tzdata \
    busybox-suid \
    && pip3 install --break-system-packages yt-dlp \
    && rm -rf /var/cache/apk/*

# Create necessary directories
RUN mkdir -p /app/config /app/downloads /app/archives /var/log/cron && \
    touch /var/log/cron/cron.log

COPY --from=builder /build/auto-yt-dlp /app/
RUN chmod +x /app/auto-yt-dlp

COPY start.sh /start.sh
RUN chmod +x /start.sh

WORKDIR /app

ENV TZ=UTC
ENV CRON_SCHEDULE="0 * * * *"

VOLUME ["/app/downloads", "/app/archives"]

CMD ["/start.sh"]

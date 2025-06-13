#!/bin/sh

CRON_SCHEDULE=${CRON_SCHEDULE:-"0 * * * *"}

echo "Creating crontab with schedule: $CRON_SCHEDULE"
cat > /tmp/crontab << EOF
$CRON_SCHEDULE cd /app && /app/auto-yt-dlp
EOF

# Install crontab for current user (root)
crontab /tmp/crontab

echo "$(date): YouTube Downloader container started"
echo "$(date): Cron schedule set to: $CRON_SCHEDULE"

echo "$(date): Starting cron daemon in foreground..."
crond -f -L /dev/stdout &
CRON_PID=$!

echo "$(date): Running initial download..."
cd /app && /app/auto-yt-dlp

echo "$(date): Container ready. Waiting..."

wait $CRON_PID


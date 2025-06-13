Sure! Here’s a complete updated **README.md** file for your **auto-yt-dlp** project, incorporating the Docker Compose example and usage instructions:

````markdown
# auto-yt-dlp

**auto-yt-dlp** is a lightweight, automated YouTube downloader built with Go and yt-dlp. It supports scheduled downloads using cron jobs inside a Docker container, allowing you to easily automate video downloads on a flexible schedule.

---

## Features

- Scheduled downloads via cron with customizable intervals
- Automatic handling of video metadata, thumbnails, and embedding
- Runs inside a minimal Docker container for easy deployment
- Supports configuration via a YAML file
- Logs output to stdout for easy container log management

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed
- (Optional) [Docker Compose](https://docs.docker.com/compose/install/) for multi-service setups

---

## Configuration

Create a `config.yaml` to customize download settings. Mount this file into the container at `/app/config/config.yaml`.

### Example `config.yaml`

```yaml
subscriptions:
  - name: "I Tried the Top 50 Airbnbs in America by Ryan Trahan"
    url: "https://www.youtube.com/@ryan/videos"
    destination: ryan_trahan
    max_videos: 10
    filter: "I Tried the Top 50 Airbnbs in America"
```

---

## Docker

### Build and Run with Docker

Build the image and run the container:

```bash
docker build -t auto-yt-dlp .
docker run -d \
  --name auto-yt-dlp \
  -e TZ=Europe/Lisbon \
  -e CRON_SCHEDULE="0 * * * *" \
  -v $(pwd)/config.yaml:/app/config/config.yaml \
  -v $(pwd)/downloads:/app/downloads \
  -v $(pwd)/archives:/app/archives \
  auto-yt-dlp
```
````

---

### Docker Compose

Alternatively, use the following `docker-compose.yml` to manage your container:

```yaml
version: "3.8"

services:
  auto-yt-dlp:
    build: .
    container_name: auto-yt-dlp
    restart: unless-stopped
    environment:
      - TZ=Europe/Lisbon
      - CRON_SCHEDULE=0 * * * * # Every hour (default)
    volumes:
      - ./config.yaml:/app/config/config.yaml # Configuration file
      - ./archives:/app/archives # Directory for archives
      - ./downloads:/app/downloads # Directory for output files
```

To start the service:

```bash
docker-compose up -d
```

View logs:

```bash
docker-compose logs -f auto-yt-dlp
```

---

## Environment Variables

- `TZ`: Time zone (default: `UTC`)
- `CRON_SCHEDULE`: Cron expression for scheduling downloads (default: `0 * * * *` — every hour)

---

## Logs

Logs are output to stdout and can be viewed using Docker’s logging commands:

```bash
docker logs -f auto-yt-dlp
```

---

## License

MIT License

---

## Contributions

Contributions and issues are welcome! Feel free to submit pull requests or open issues.

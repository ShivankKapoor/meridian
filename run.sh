#!/bin/bash
set -e

mkdir -p logs

TIMESTAMP=$(TZ="America/Chicago" date +"%Y-%m-%d_%I-%M-%S_%p_CST")
LOG_FILE="$(pwd)/logs/meridian_${TIMESTAMP}.log"

podman build -t meridian .
podman run --rm --name Meridian -p 9090:9090 -v "$(pwd)/.env:/app/.env:z" -d meridian

podman logs -f Meridian >> "$LOG_FILE" 2>&1 &
echo "Container started. Logs writing to $LOG_FILE"

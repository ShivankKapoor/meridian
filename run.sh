#!/bin/bash
set -e

podman build -t meridian .
podman run --rm --name Meridian -p 9090:9090 -v "$(pwd)/.env:/app/.env:z" -d meridian

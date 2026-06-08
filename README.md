# Meridian

A fast IP geolocation API written in Go. Looks up location data for a given IP address, caches results in Redis, and falls back to [ip-api.com](http://ip-api.com) on a cache miss.

## Features

- IP geolocation via ip-api.com
- Redis caching with a 7-day TTL
- IPv4 and IPv6 support
- Discord webhook alerts on cache misses and Redis outages
- Graceful shutdown on SIGTERM/SIGINT

## Requirements

- [Podman](https://podman.io/)
- Redis instance
- Discord webhook URL (optional)

## Setup

Copy `example.env` to `.env` and fill in your values:

```env
PORT=9090
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
DISCORD_WEBHOOK=
```

## Running

```bash
./run.sh
```

Builds the image, starts the container, and writes logs to `logs/meridian_<timestamp>.log`.

```bash
./stop.sh
```

Stops the container and removes the image.

## API

### `GET /location/{ip}`

Returns geolocation data for the given IP address.

**Request**
```
GET /location/8.8.8.8
```

**Response**
```json
{
  "country": "United States",
  "countryCode": "US",
  "city": "Ashburn",
  "regionName": "Virginia"
}
```

**Errors**

| Status | Reason |
|--------|--------|
| `400`  | Invalid IP address |
| `500`  | Failed to fetch location |

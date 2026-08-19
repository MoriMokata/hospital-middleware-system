# Docker

- `Dockerfile` — multi-stage build of the Go API service (build stage compiles a static binary; runtime stage is a minimal Alpine image running as a non-root user).
- `nginx/nginx.conf` — reverse proxy in front of the `api` service, with basic per-IP rate limiting and a dedicated `/health` passthrough.

The root [`../docker-compose.yml`](../docker-compose.yml) wires `nginx`, `api`, and `postgres` together on an internal network — only `nginx` is published to the host.

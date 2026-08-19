# ---- build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

# ---- runtime stage ----
FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/api ./api
USER app
EXPOSE 8080
ENTRYPOINT ["./api"]

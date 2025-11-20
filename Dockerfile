# syntax=docker/dockerfile:1
ARG GO_VERSION=1.23.0
FROM golang:${GO_VERSION}-bullseye AS build
WORKDIR /src

# Copy the Go modules files first and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application code
COPY . .

# Build the application without cross-compiling
RUN CGO_ENABLED=1 go build -o /base-api .

FROM debian:bookworm-slim AS final
WORKDIR /app

# Install certificates and time zone data
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

ARG UID=10001
RUN useradd -r -u ${UID} appuser

# Copy the binary and set permissions before switching users
COPY --from=build /base-api .
RUN chown appuser:appuser /app/base-api
RUN chmod +x /app/base-api

# Copy other files as needed
COPY --from=build /src/logs ./logs
COPY --from=build /src/storage ./storage

RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

EXPOSE 8091
ENTRYPOINT [ "./base-api" ]
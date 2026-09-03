# Multi-stage build: compile the static-ish binary, ship it in a minimal
# runtime image with just libseccomp. The container still needs to run
# --privileged (or with the right namespace capabilities) for offline itself
# to create namespaces, so this mainly exists for CI artifact builds.

# --- build stage ---
FROM golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437 AS build
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends pkg-config libseccomp-dev \
 && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY offline.go ./
RUN CGO_ENABLED=1 go build -o offline offline.go

# --- runtime stage ---
FROM debian:bookworm-slim@sha256:88200866dfff7ea7f5cbcb6ec7c8a701889efe6fe859fe64d6990e4b07ea4171
RUN apt-get update && apt-get install -y --no-install-recommends libseccomp2 \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /src/offline /app/offline
ENTRYPOINT ["/app/offline"]

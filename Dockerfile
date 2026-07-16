# Multi-stage build: compile the static-ish binary, ship it in a minimal
# runtime image with just libseccomp. The container still needs to run
# --privileged (or with the right namespace capabilities) for offline itself
# to create namespaces, so this mainly exists for CI artifact builds.

# --- build stage ---
FROM golang:1.25-bookworm AS build
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends pkg-config libseccomp-dev \
 && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY offline.go ./
RUN CGO_ENABLED=1 go build -o offline offline.go

# --- runtime stage ---
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends libseccomp2 \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /src/offline /app/offline
ENTRYPOINT ["/app/offline"]

# Multi-stage build: compile the static-ish binary, ship it in a minimal
# runtime image with just libseccomp. The container still needs to run
# --privileged (or with the right namespace capabilities) for offline itself
# to create namespaces, so this mainly exists for CI artifact builds.

# --- build stage ---
FROM golang:1.27-bookworm@sha256:648f440f42a0958804efb24df176f806f9d353b41f1c0627f666428e40310f6b AS build
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends pkg-config libseccomp-dev \
 && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY offline.go ./
RUN CGO_ENABLED=1 go build -o offline offline.go

# --- runtime stage ---
FROM debian:bookworm-slim@sha256:3783cc01769c7b2b1b83a5c5ad96c815348e28ed7da68e2e3687004faa906251
RUN apt-get update && apt-get install -y --no-install-recommends libseccomp2 \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /src/offline /app/offline
# hardener: run this image with `docker run --read-only` for a read-only rootfs
# The hardener's other suggestion, `USER 10001`, is deliberately not applied:
# offline creates user/net/mount/PID/IPC/UTS namespaces, this image is
# documented as needing --privileged to do so, and CLAUDE.md treats the
# privilege context as a security decision rather than a routine hardening.
ENTRYPOINT ["/app/offline"]

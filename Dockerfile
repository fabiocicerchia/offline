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
FROM debian:bookworm-slim@sha256:abd67ffcfa541b485a3dff59865ab629aa048a6c613e639d36e7456b0b229241
RUN apt-get update && apt-get install -y --no-install-recommends libseccomp2 \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /src/offline /app/offline
ENTRYPOINT ["/app/offline"]

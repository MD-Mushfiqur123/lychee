# =============================================================================
# Docker Multi-Stage Build for Lychee (with full CGO support)
# =============================================================================
#
# CGO is required because Lychee embeds llama.cpp (a C++ library) for GPU-
# accelerated local LLM inference. Without CGO, the Go bindings to llama.cpp
# cannot be compiled, and the binary will lack model loading/running capability.
#
# Build requirements:
#   - gcc / g++          : C and C++ compilers (needed for CGO + llama.cpp)
#   - musl-dev           : musl C library headers (Alpine's libc, required by CGO)
#   - cmake / make       : build system for llama.cpp's CMake-based compilation
#   - linux-headers      : kernel headers needed by certain CGO packages
# =============================================================================

# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

# Install C/C++ build toolchain for CGO and llama.cpp compilation
RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    cmake \
    make \
    linux-headers

WORKDIR /build

# Cache Go modules independently to speed up rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build with CGO enabled (required for llama.cpp bindings)
# -ldflags="-s -w" strips debug symbols to reduce binary size
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /lychee .

# ---- Runtime Stage ----
FROM alpine:latest

# Install CA certificates for HTTPS connections (e.g., model downloads, API calls)
RUN apk add --no-cache ca-certificates

# Copy the statically-linked binary from the build stage
COPY --from=builder /lychee /usr/local/bin/lychee

# Lychee's default server port
EXPOSE 11434

# Start the Lychee server
CMD ["lychee", "serve"]

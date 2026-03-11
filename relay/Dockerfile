# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache dependency downloads separately from the source build.
COPY go.mod go.sum ./
# The relay module replaces ../protocol via a local replace directive.
# Copy the sibling protocol module so the build can resolve it.
COPY ../protocol /protocol

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /relay ./cmd/relay/

# ── Final stage ───────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

# nonroot user (65534) is baked into the distroless:nonroot image.
USER 65534

COPY --from=builder /relay /relay

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/relay", "-healthcheck"]

# Use exec form so signals are delivered directly to the binary.
ENTRYPOINT ["/relay"]

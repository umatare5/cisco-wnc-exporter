# Dockerfile for cisco-wnc-exporter

FROM scratch

# Copy ca-certificates for HTTPS requests to cisco-wnc-exporter controllers
COPY --from=alpine:latest@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the pre-built binary from GoReleaser
COPY cisco-wnc-exporter /cisco-wnc-exporter

# Create a non-root user (using numeric ID for scratch image)
USER 65534:65534

# Set the entrypoint
ENTRYPOINT ["/cisco-wnc-exporter"]

# Default command shows help
CMD ["--help"]

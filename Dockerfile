# Dockerfile for cisco-wnc-exporter

FROM scratch

# dockers_v2 lays the build context out as linux/<arch>/<binary>
ARG TARGETPLATFORM

# Copy ca-certificates for HTTPS requests to cisco-wnc-exporter controllers
COPY --from=alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the pre-built binary from GoReleaser
COPY $TARGETPLATFORM/cisco-wnc-exporter /cisco-wnc-exporter

# extra_files in .goreleaser.yml is what puts these in the build context
COPY LICENSE NOTICE /

# Create a non-root user (using numeric ID for scratch image)
USER 65534:65534

# Declare the port; publishing it still requires docker run -p
EXPOSE 10039

# Set the entrypoint
ENTRYPOINT ["/cisco-wnc-exporter"]

# Security Policy

## Supported versions

This project is pre-1.0, so only the latest release carries fixes. A security fix lands on `main` and ships in the next release; there is no patch branch for an older tag. Reproduce a finding against the latest release before reporting it.

## Reporting a vulnerability

Report privately through GitHub Security Advisories: open the repository's **Security** tab and choose **Report a vulnerability**. Do not open an issue or a pull request for a suspected vulnerability, because both are world-readable from the moment they are created.

This project is maintained by one person in their own time, so no response time is promised. A fix is released before the advisory is published, and the advisory credits the reporter unless they ask otherwise.

## What this exporter holds and exposes

- It holds a controller credential for its whole lifetime. `--wnc.access-token` is the Base64 of `username:password` for a controller account, so the exporter is exactly as sensitive as that account. Prefer `WNC_ACCESS_TOKEN` over the flag, which puts the credential in the process table.
- `--wnc.tls-skip-verify` disables certificate verification on the controller connection, which leaves the read open to interception. It exists for a controller with a self-signed certificate and is not for production — see the warning in [README.md](README.md).
- `/metrics` is served over plain HTTP with no authentication. Its series carry AP and client MAC addresses, and the default `--collector.*.info-labels` sets already publish the AP name, the AP IP address and the client IPv4 address; the client username and IPv6 address are the labels that are opt-in. Reach the endpoint over a controlled path rather than an exposed one.
- A controller response with a status of 400 or above has its body written to the log at error level, and the same body is carried in the error string this exporter logs in turn. An error response body is the controller's own error document rather than configuration data, but nothing truncates or redacts it. The credential does not appear in either line.

## Out of scope

A defect in the controller's own RESTCONF interface is not this project's to fix. This is a third-party exporter, so report such a finding to the vendor.

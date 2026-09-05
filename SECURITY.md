# Security Policy

## Supported versions

Only the latest release carries fixes, and no older tag gets a patch branch. Reproduce a finding against the latest release before reporting it.

## Reporting a vulnerability

Report privately through GitHub Security Advisories, never an issue or a pull request — open the repository's **Security** tab and choose **Report a vulnerability**.

One maintainer works on this in their own time, so no response time is promised. The advisory goes out after the fix ships and credits the reporter unless they ask otherwise.

## What this exporter holds and exposes

This exporter uses a controller account to read the controller's RESTCONF interface, and exposes the data as Prometheus metrics. The account's privileges determine what the exporter can read, so a read-only account is enough for all metrics.

- **Credential** — one controller account, which makes the exporter exactly as sensitive as that account, so prefer `WNC_ACCESS_TOKEN` to `--wnc.access-token`.
- **Certificate verification** — `--wnc.tls-skip-verify` skips it for a self-signed certificate and is not for production, as the caution in [`README.md`](README.md) says.
- **Metrics** — unauthenticated plain HTTP carrying AP and client MAC addresses and the default [`--collector.*.info-labels`](docs/help.md) sets, so keep it on a controlled path.
- **Logs** — a controller response of 400 or above reaches the log at error level unredacted, and the SDK bounds it to a prefix that carries no credential.

## Out of scope

A defect in the controller's own RESTCONF interface belongs to the **Cisco** — report it there, not to this third-party exporter.

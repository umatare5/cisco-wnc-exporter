# Security Policy

The [shared policy](https://github.com/umatare5/.github/blob/main/SECURITY.md) covers what every exporter here shares. This page carries the rest.

## What to Include

Redact these before reporting, in addition to the credentials the shared policy names.

- The `WNC_ACCESS_TOKEN` value, which is the account and its password in reversible base64
- An AP or client MAC address, an AP name, a client username or an ESSID, from `/metrics`
- The controller's address, from `--wnc.controller`, an error log line or a capture

Reproduction needs the collector module flags in force and the controller's IOS-XE version.

## Exposure

This exporter reads a Cisco Catalyst 9800 over RESTCONF and republishes the readings as Prometheus metrics. Every request it issues is a GET, so the account's privileges bound the whole surface and a read-only account serves every collector.

- **Credential** — `WNC_ACCESS_TOKEN` is the base64 of `user:password` the SDK sends as `Authorization: Basic`, so it is a reversible encoding of the account password rather than a scoped token.
- **Flag over environment** — `--wnc.access-token` puts that same value on the command line, where every account on the host reads it out of `ps`, so pass the environment variable instead.
- **Certificate verification** — `--wnc.tls-skip-verify` drops it, which hands the credential to whatever answers for the controller's address, so it belongs to a lab rather than to production.
- **Metrics** — AP and client MAC addresses identify their series whatever the flags, and the default `--collector.*.info-labels` add the AP name and IP, the client name and IPv4, and the ESSID.
- **Opt-in labels** — `username`, `ipv6` and `device_type` reach `/metrics` only where the flag names them, so the label set an operator chooses decides how much of a client the endpoint discloses.
- **Logs** — a controller response of 400 or above reaches the log unredacted, bounded by the SDK to 512 bytes of a rejection document that is a response body and so holds no credential.

> [!IMPORTANT]
> An account with write privilege collects nothing a read-only account does not, because no code path builds anything but a GET, so granting one widens only what a leaked credential reaches.

## Out of Scope

A defect in the controller's own RESTCONF implementation belongs to Cisco.

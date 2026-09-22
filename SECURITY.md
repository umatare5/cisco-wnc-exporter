# Security Policy

Please follow **[the shared security policy](https://github.com/umatare5/.github/blob/main/SECURITY.md)**, which covers:

- **Supported Versions** – only the latest release carries fixes, so reproduce against it.
- **Reporting a Vulnerability** – the private advisory path and what the response promises.
- **What to Include** – the credentials and addresses to redact, and the fields to send.
- **Exposure** – the unauthenticated surface and the container the image ships.
- **Out of Scope** – findings that belong to the monitored system or to an operator's own configuration.

This page specifies what is particular to this one.

## What to Include

Reproduction needs the collector flags in force and the controller's IOS-XE version.

Redact these before reporting, in addition to the credentials the shared policy names.

- **Infrastructure** – The controller's address, from `--wnc.controller`, an error log line or a capture.
- **Application** – An AP or client MAC address, an AP name, a client username or an ESSID, from `/metrics`.
- **Credentials** – The `WNC_ACCESS_TOKEN` value, which is the account and its password in reversible base64.

## Exposure

This exporter reads a Cisco Catalyst 9800 over RESTCONF and republishes the readings as Prometheus metrics. Every request it issues is a GET, so the account's privileges bound the whole surface and a read-only account serves every collector.

- **Metrics** – AP and client MAC addresses identify their series whatever the flags.
- **Info labels** – the default `--collector.*.info-labels` add the AP name and IP.
- **The same defaults** – they add the client name and IPv4, and the WLAN name on `wnc_wlan_info`.
- **Opt-in labels** – `username`, `ipv6` and `device_type` reach `/metrics` only where a flag names them.
- **Disclosure** – the label set an operator chooses decides how much of a client the endpoint shows.
- **Unauthenticated** – `/metrics` authenticates nobody, so a network path is the only control.
- **Logs** – a controller response of 400 or above reaches the log unredacted.
- **Bounded** – the SDK caps that line at 512 bytes of the rejection document.
- **No credential** – that document is a response body, so it holds none.

> [!IMPORTANT]
> An account with write privilege collects nothing a read-only account does not, because no code path builds anything but a GET, so granting one widens only what a leaked credential reaches.

## Endpoints

No route authenticates, so the network path the listener sits on is the whole access control. See also [Endpoints](docs/architecture.md#endpoints).

- **Listener** – `--web.listen-address` defaults to `0.0.0.0`, which answers on every interface.
- **Metrics** – `/metrics` serves the aggregated data.
- **Health** – `/healthz` provides a static `OK` liveness probe.
- **Restriction** – put a packet filter or an authenticating proxy in front of the metrics endpoint.

## Ingress Paths

The exporter exposes a listening socket for incoming HTTP scrapes.

- **Reach** – `--web.listen-address` controls which interfaces accept connections.
- **No allowlist** – the exporter filters no sender, which leaves the packet filter to enforce it.

## Egress Paths

The exporter opens one outbound connection, an HTTPS RESTCONF session to `--wnc.controller`.

- **Credential** – `WNC_ACCESS_TOKEN` is the base64 of `user:password`, sent as `Authorization: Basic`.
- **Not a scoped token** – that encoding is reversible, so it carries the account password itself.
- **Flag over environment** – `--wnc.access-token` puts that same value on the command line.
- **Every account reads it** – the line is in `ps`, so pass the environment variable instead.
- **Certificate verification** – `--wnc.tls-skip-verify` drops it, which belongs to a lab.
- **Production** – dropping it hands the credential to whatever answers for that address.

> [!NOTE]
> The exporter initiates no other connection: none for telemetry, updates or error reporting, and it consults no proxy variable. The controller – plus the resolver, where `--wnc.controller` names a host rather than an address – is the whole outbound surface.

## Out of Scope

- **Origin** – A defect in the controller's own RESTCONF implementation belongs to Cisco rather than to this exporter.

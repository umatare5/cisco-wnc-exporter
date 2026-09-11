# Contributing

The [shared contribution guide](https://github.com/umatare5/.github/blob/main/CONTRIBUTING.md) covers what every exporter shares. This page carries the rest.

## Development

CI runs Format and Lint, Test and Build, Coverage against a threshold of 80 percent, Prometheus Rules and CodeQL on every pull request. The markdownlint, Link Check, actionlint and govulncheck jobs are gated on the paths they read — Markdown, workflows, Go sources and the module files among them. A change touching none of those paths skips all four.

## Testing

Every collector test reads one controller snapshot rather than a fixture of its own, because the collectors share a cache and a leaf added to one of them changes what the others withhold.

- **The snapshot** — `internal/collector/absence_test.go` holds `fullFixtureSnapshot`, the whole reading.
- **One place** — a new leaf is added there once, and every collector test sees it.
- **Absence is asserted** — a data type marked failed publishes no series the healthy gather did not.
- **What it pins** — a failure belongs to the types that failed.
- **`promlint` covers the surface** — it lints every gathered family.
- **No list of regressions** — a counter that loses its `_total` suffix fails without one kept by hand.

Three commands reproduce the `Prometheus Rules` job locally.

```bash
promtool check rules --lint all --lint-fatal examples/prometheus_alert_rules.yml
promtool test rules examples/prometheus_alert_rules_test.yml
promtool check config --lint all --lint-fatal examples/prometheus.yml
```

A value reaches a fixture or a HELP string only after a controller sent it.

```bash
# A collection
curl -k -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
     -H "Accept: application/yang-data+json" \
     "https://$WNC_CONTROLLER/restconf/data/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data"

# One entry, keyed by MAC address
curl -k -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
     -H "Accept: application/yang-data+json" \
     "https://$WNC_CONTROLLER/restconf/data/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data=00:11:22:33:44:55"

# An RPC, which this exporter never issues and which changes the controller
curl -k -X POST -H "Authorization: Basic $WNC_ACCESS_TOKEN" \
     -H "Content-Type: application/yang-data+json" \
     -d '{"input": {"ap-name": "TEST-AP01"}}' \
     "https://$WNC_CONTROLLER/restconf/operations/Cisco-IOS-XE-wireless-access-point-cmd-rpc:ap-reset"
```

> [!CAUTION]
> Append `?with-defaults=report-all` to a configuration read to see the leaves a default hides. Never append it to an operational read, because materialising defaults there defeats the guards that keep a fabricated zero out of the metrics.

## Code Style

A `--collector.<name>.<group>` flag switches one group of families inside one of the four service collectors, and the two `--collector.internal.*` flags sit outside them, registering the exporter's own Go runtime and process families. No bare collector flag exists, so a collector publishes nothing until one of its group flags is set.

A HELP string states the reading of one series in one sentence, and it says whether the series goes absent or decodes an omitted leaf as `0`, because a C9800 omits a leaf holding its schema default.

## Documentation

Every fact has one page that owns it, and the other pages link to it rather than restating it.

| Page                  | Owns                                        |
| :-------------------- | :------------------------------------------ |
| `README.md`           | What it is, how to run and scrape it        |
| `docs/README.md`      | The rules every collector obeys             |
| `docs/collector.*.md` | The metric catalogue of one collector       |
| `docs/health.md`      | The exporter's own build and refresh series |
| `docs/enums.md`       | The number each enumerated family reports   |
| `docs/help.md`        | The verbatim `--help` transcript            |

> [!NOTE]
> A sentence about what a leaf carries is written after that leaf was read off a controller, and `docs/enums.md` names the YANG module and revision date each enumeration was taken from.

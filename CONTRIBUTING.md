# Contributing

Thank you for considering a contribution.

## Commands

The following `make` commands are available for development and testing:

| Command                   | Description                                    |
| :------------------------ | :--------------------------------------------- |
| `make help`               | Display available targets and requirements     |
| `make build`              | Build the binary to `./tmp/cisco-wnc-exporter` |
| `make lint`               | Run golangci-lint and tidy go.mod              |
| `make test-unit`          | Run unit tests with coverage using gotestsum   |
| `make test-unit-coverage` | Generate HTML coverage report                  |
| `make clean`              | Remove build artifacts and backup files        |
| `make image`              | Build Docker image                             |

## Build

The repository includes a ready to use `Dockerfile`. To build a new Docker image:

```bash
make image
```

This cross-compiles a Linux binary into `./tmp/image`, then builds from that directory because the `Dockerfile` expects the binary at the context root. The image is tagged `$USER/cisco-wnc-exporter` and declares port 10039 without publishing it, so publish it with `docker run -p`. Released images are pushed to `ghcr.io/umatare5/cisco-wnc-exporter` by GoReleaser instead.

## Release

To release a new version, follow these steps:

1. Rename the `## Unreleased` heading in `CHANGELOG.md` to `## vX.Y.Z`, matching the `VERSION` file exactly.
2. Update the version in the `VERSION` file.
3. Update the `VERSION:` line in the `--help` transcript in `docs/configuration.md`.
4. Submit a pull request with all three files.

Merging that pull request is the whole release. A push to `main` touching `VERSION` runs the [release workflow](https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-release.yml), which tags the commit and publishes the release in the same run. The workflow has no manual trigger, so there is no step to perform by hand.

## Pull requests

1. Fork ([https://github.com/umatare5/cisco-wnc-exporter/fork](https://github.com/umatare5/cisco-wnc-exporter/fork))
2. Create a feature branch
3. Commit your changes
4. Record any change to the metric surface under `## Unreleased` in `CHANGELOG.md`
5. Rebase your local changes against the `main` branch
6. Create a new Pull Request

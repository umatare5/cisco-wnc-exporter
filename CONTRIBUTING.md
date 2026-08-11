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

This creates an image named `ghcr.io/$USER/cisco-wnc-exporter` and exposes `10039/tcp`.

## Release

To release a new version, follow these steps:

1. Update the version in the `VERSION` file.
2. Submit a pull request with the updated `VERSION` file.

Once the pull request is merged, the GitHub Workflow automatically creates and pushes a new tag, after which I manually publish a release using the [GitHub Actions release workflow](https://github.com/umatare5/cisco-wnc-exporter/actions/workflows/go-release.yml).

## Pull requests

1. Fork ([https://github.com/umatare5/cisco-wnc-exporter/fork](https://github.com/umatare5/cisco-wnc-exporter/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Create a new Pull Request

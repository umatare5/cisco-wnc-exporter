# Contributing

Thank you for your interest in contributing to the cisco-wnc-exporter.

Please follow **[the shared contribution guide](https://github.com/umatare5/.github/blob/main/CONTRIBUTING.md)**, which covers:

- **Development** – the tools to install and the order the pre-commit hooks run in.
- **Command** – the `make` targets and what each one does.
- **Testing** – test placement, mutation checks and what a fixture must carry.
- **Documentation** – page ownership, pinned headings and the verbatim `--help` transcript.
- **Release** – the three files a release touches and what a push to `main` triggers.
- **Pull Requests** – the branch, commit and changelog steps, and what never enters a commit.

This page specifies what is particular to this one.

## Development

These points are where this repository departs from the shared defaults.

- **Do not assume every check runs.** Four are path-filtered: govulncheck, markdownlint, Link Check and actionlint.
- **Keep coverage above 80 percent.** `make test-unit` writes the profile, and the coverage workflow is what judges it.
- **Clear an actionlint finding before pushing.** The shared workflow lets a finding fail the step rather than exiting zero.
- **Do not read `make lint` as the CI gate.** CI pins the linter to the version the caller names and runs `go mod verify` first.
- **Stamp the version with `make build`.** The `--help` transcript reads `dev` until the build stamps it.

## Testing

These points are where this repository's tests depart from the shared approach.

- **Extend `fullFixtureSnapshot` rather than adding a fixture.** Every collector test reads that one controller snapshot.
- **Assert absence rather than presence alone.** A data type marked as failed must publish no series the healthy gather did not.
- **Keep a new family `promlint`-clean.** One test lints every gathered family rather than a sampled few.

## Documentation

The shared guide defines one owner per fact, and [Documentation](README.md#documentation) names the owner of each.

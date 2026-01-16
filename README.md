# hashi

> Hashing utilities and experiments — simple, fast, and configurable.

## Overview

Hash Machine is a small Go project that provides hashing utilities and examples for computing and comparing hashes. This repository contains the CLI and library code to build, test, and run small hashing workflows.

## Requirements

- Go 1.20+ (set in `go.mod`)

## Build

To build a local binary:

```bash
go build -o hashi
```

Run tests:

```bash
go test ./...
```

## Usage

Example (after building):

```bash
./hashi [flags]
# see `--help` for available commands
```

## Contributing

Contributions are welcome. Please open issues for bugs and feature requests. For code changes, open a pull request with tests and a short description of the change.

## License

This project is unlicensed by default. Add a `LICENSE` file if you want to apply a specific license.

## Releases

This repository is being released as `hashi 1.0`.

- Tag: `v1.0.0`
- Release notes: see `CHANGELOG.md` for details

To publish the release (example using the GitHub CLI):

```bash
# push commits and tags
git push origin HEAD
git push origin v1.0.0

# create a GitHub release from the tag
gh release create v1.0.0 --title "Hash Machine v1.0.0" --notes-file CHANGELOG.md
```

If you want release assets (binaries), build them for target platforms and pass `--assets` to `gh release create`.

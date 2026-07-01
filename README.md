# Terraform Provider Packager

The Terraform Provider Packager is a utility designed to prepare and structure Terraform provider assets for use in a static asset-based Terraform registry.

This tool automates the process of:
 - Building the required directory structure for provider distribution
 - Copying and organizing provider binaries and metadata
 - Ensuring compatibility with Terraform’s provider installation protocol

It streamlines packaging providers for offline use or for serving via a custom/private Terraform provider registry.

## Features
 - 📁 Creates Terraform-compliant registry paths
 - 🗃 Copies provider binaries and SHA256SUMS metadata
 - ✅ Supports versioned provider releases
 - 🧩 Easy integration into CI/CD workflows

## Use Case

This packager is ideal when hosting Terraform providers in:
 - A private or air-gapped environment
 - An internal HTTP server
 - A static hosting service (e.g., S3, GitHub Pages)

## Getting Started

### Create Gorelease dist
```bash
goreleaser release --clean
```

### Create package from Goreleaser

```bash
tfpp -p example -r terraform-provider-example  \
-ns=exampleorg \
-d=terraform-registry.example.com \
-gf=$GPG_FINGERPRINT \
-v=1.0.0
```

### Copy to S3

```bash
aws s3 sync release/ s3://s3-tfregistry-example/
```

## GitHub Action

`tfpp` ships as a composite GitHub Action so you can package a provider directly
in your release pipeline without installing the binary yourself. The action
downloads the prebuilt `tfpp` binary matching the runner OS/arch, verifies its
checksum, and runs it against your GoReleaser `dist/` directory.

### Usage

```yaml
- name: Import GPG key
  id: import_gpg
  uses: crazy-max/ghaction-import-gpg@v7
  with:
    gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
    passphrase: ${{ secrets.PASSPHRASE }}

- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v7
  with:
    args: release --clean
  env:
    GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}

- name: Package for the static registry
  id: package
  uses: marceloalmeida/tfpp@v1
  with:
    namespace: exampleorg
    domain: terraform-registry.example.com
    provider-name: example
    repo-name: terraform-provider-example
    version: 1.0.0
    gpg-fingerprint: ${{ steps.import_gpg.outputs.fingerprint }}

- name: Publish to S3
  run: aws s3 sync "${{ steps.package.outputs.release-path }}/" s3://s3-tfregistry-example/
```

Pin to a specific tool release with `tfpp-version: 1.2.3` for reproducible builds
(it defaults to the latest release).

### Inputs

| Input | Required | Default | tfpp flag | Description |
| --- | --- | --- | --- | --- |
| `namespace` | yes | | `-ns` | Namespace for the Terraform registry. |
| `domain` | yes | | `-d` | Private Terraform registry domain. |
| `provider-name` | yes | | `-p` | Name of the Terraform provider. |
| `repo-name` | yes | | `-r` | Provider repository name used in the GoReleaser build files. |
| `version` | yes | | `-v` | Semantic version of the **provider release** being packaged. |
| `gpg-fingerprint` | yes | | `-gf` | GPG fingerprint of the key used by GoReleaser. |
| `dist-path` | no | `dist` | `-dp` | Path to the GoReleaser build files. |
| `gpg-public-key-file` | no | `pubkey.txt` | `-gk` | Path to the GPG public key in ASCII armor format. |
| `tfpp-version` | no | `latest` | | Version of the **tfpp tool** to download (e.g. `1.2.3`). |
| `working-directory` | no | `.` | | Directory to run `tfpp` in; `release/` is created here. |
| `github-token` | no | `${{ github.token }}` | | Token used to query the tfpp releases API. |

> Note: `version` is the provider version you are packaging, while `tfpp-version`
> selects which build of the `tfpp` tool the action downloads.

### Outputs

| Output | Description |
| --- | --- |
| `release-path` | Absolute path to the generated `release/` directory. |

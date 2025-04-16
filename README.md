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

## TODO
- [ ] Provide a Github Actions workflow example
- [ ] Add unit tests

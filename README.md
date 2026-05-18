# gosho

<p align="center">
  <img src="assets/gosho.png" alt="gosho" width="200">
</p>

An interactive AWS SSO login CLI that forces fresh browser sessions (InPrivate/Incognito) per login, solving the problem of cached SAML credentials when switching between environments (e.g., DEV vs PRD) under the same SSO start URL.

## Why

`aws sso login` reuses the browser session cache, so when you need different SAML credentials for different environment (e.g DEV vs PRD), the cached browser session may interfere with the profile selection. Gosho solves this by openibg an InPrivate/Incognito window each time, ensuring you're prompted for the correct credentials.

## Install

Requires Go 1.25+.

```bash
# Build
make build

# Install to ~/.local/bin
make install
```

Or directly:

```bash
go install github.com/gandhinn/gosho@latest
```

## Usage

```bash
gosho
```

The interactive flow:

1. Prompts for SSO start URL and region
2. Opens an InPrivate browser window for device authorization
3. Lists available accounts -> select one
4. Lists available roles -> select one
5. Prompts for a profile name
6. Writes credentials to `~/.aws/credentials` under that profile
7. Prompts you to close the browser before logging into another environment

Then use the profile with any AWS tool:

```bash
aws s3 ls --profile icloud-dev
sift security --profile icloud-prd
```

## How it works

- Registers an OIDC device client with AWS SSO
- Opens Edge InPrivate (WSL) or Incognito (Linux/macOS) to avoid session reuse
- Polls for token completion
- Retrieves role credentials and writes them as static credentials
- Caches tokens per profile under `~/.gosho/cache/`

## Project structure

```bash
gosho/
├── main.go              # Entry point
├── Makefile             # Build/install
├── cmd/
│   └── login.go         # Interactive login flow
├── sso/
│   ├── constant.go      # Regions, grant types, scopes
│   ├── auth.go          # OIDC device auth, browser launch
│   ├── token.go         # Token cache (per profile)
│   ├── account.go       # List accounts/roles, get credentials
│   └── credentials.go   # Write to ~/.aws/credentials
└── config/
    └── config.go        # (Optional) YAML config for presets
```

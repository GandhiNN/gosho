# gosho

<p align="center">
  <img src="assets/gosho.png" alt="gosho" width="200">
</p>

AWS SSO login CLI with fresh browser sessions for clean multi-environment authentication.

## Why

`aws sso login` reuses the browser session cache, so when you need different SAML credentials for different environments (e.g., DEV vs PRD), the cached browser session interferes. Gosho opens an InPrivate/Incognito window each time, ensuring you're prompted for the correct credentials.

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
# First time: interactive (saves preset for future use)
gosho

# With saved preset: skips account/role selection
gosho icloud-dev

# Configure defaults (start URL, region)
gosho init

# Check cached profile status
gosho status
```

### Interactive flow

1. Prompts for SSO start URL and region (or uses saved defaults)
2. Opens an InPrivate/Incognito browser window for device authorization
3. Lists available accounts → select one
4. Lists available roles → select one
5. Prompts for a profile name
6. Writes credentials to `~/.aws/credentials` under that profile
7. Saves the account/role as a preset in `~/.gosho/config.yaml`
8. Prompts you to close the browser before logging into another environment

### Preset flow

Once a profile has been used interactively, it's saved to config. Subsequent runs skip account/role selection:

```bash
gosho icloud-prd
# → opens fresh InPrivate browser
# → authenticates with PRD SAML credentials
# → writes credentials directly (no prompts)
```

## Configuration

Config file location: `~/.gosho/config.yaml` (override with `GOSHO_CONFIG` env var)

```yaml
start_url: https://company.awsapps.com/start
region: eu-west-1
profiles:
  icloud-dev:
    account_id: "111111111111"
    role: DevOps
  icloud-prd:
    account_id: "568650317375"
    role: DevOps
```

Profiles are saved automatically after the first interactive login.

## How it works

- Registers an OIDC device client with AWS SSO
- Opens Edge InPrivate (WSL) or Incognito (Linux/macOS) to avoid session reuse
- Polls for token completion with a spinner
- Retrieves role credentials and writes them as static credentials
- Caches tokens per profile under `~/.gosho/cache/`

## Project structure

```
gosho/
├── main.go              # Entry point
├── Makefile             # Build/install
├── cmd/
│   ├── login.go         # Interactive + preset login flow
│   ├── init.go          # Configure defaults
│   └── status.go        # Show cached profile status
├── sso/
│   ├── constant.go      # Regions, grant types, scopes
│   ├── auth.go          # OIDC device auth, browser launch, WSL detection
│   ├── token.go         # Token cache (per profile)
│   ├── account.go       # List accounts/roles, get credentials
│   └── credentials.go   # Write to ~/.aws/credentials
└── config/
    └── config.go        # YAML config (start URL, region, profile presets)
```

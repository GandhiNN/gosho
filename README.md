# gosho

<p align="center">
  <img src="assets/gosho.png" alt="gosho" width="200">
</p>

AWS SSO login CLI with fresh browser sessions for clean multi-environment authentication.

## Why

`aws sso login` reuses the browser session cache, so when you need different SAML credentials for different environments (e.g., DEV vs PRD), the cached browser session interferes. Gosho opens an InPrivate/Incognito window each time, ensuring you're prompted for the correct credentials.

## Install

Requires Go 1.25+.

### Linux / macOS

```bash
# Build
make build

# Install to ~/.local/bin
make install
```

### Windows (PowerShell)

```powershell
# Build
.\build.ps1 build

# Install to ~\.local\bin
.\build.ps1 install
```

### Go install (any platform)

```bash
go install github.com/gandhinn/gosho@latest
```

### Development setup

```bash
pre-commit install
```

## Usage

```bash
gosho login              # Interactive (saves preset for future use)
gosho login aws-dev      # Use saved preset (skips account/role selection)
gosho login all          # Login to all saved profiles in sequence
gosho logout aws-dev     # Clear cached token and credentials
gosho logout all         # Clear all cached tokens and credentials
gosho init               # Configure default start URL and region
gosho status             # Show cached profile status with expiry
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
gosho login aws-prd
# → reuses cached token if still valid (no browser)
# → or opens fresh InPrivate browser if expired
# → writes credentials directly (no prompts)
```

### Token refresh

If a cached token is still valid, gosho reuses it without opening a browser. If expired but refreshable, it refreshes automatically. A browser only opens when necessary.

## Configuration

Config file location: `~/.gosho/config.yaml` (override with `GOSHO_CONFIG` env var)

```yaml
start_url: https://company.awsapps.com/start
region: eu-west-1
profiles:
  aws-dev:
    account_id: "111111111111"
    role: DevOps
  aws-prd:
    account_id: "222222222222"
    role: DevOps
```

Profiles are saved automatically after the first interactive login.

## How it works

- Registers an OIDC device client with AWS SSO
- Opens Edge InPrivate (Windows/WSL), Chrome Incognito (Linux/macOS), or Firefox Private to avoid session reuse
- Polls for token completion with a spinner
- Caches and refreshes tokens per profile under `~/.gosho/cache/`
- Writes the SSO token to `~/.aws/sso/cache/` for AWS CLI compatibility
- Retrieves role credentials and writes them as static credentials to `~/.aws/credentials`
- Respects `AWS_SHARED_CREDENTIALS_FILE` if set
- Shows credential expiry time after each login

### AWS CLI compatibility

Gosho writes tokens to the AWS CLI's SSO cache (`~/.aws/sso/cache/`), so both the AWS CLI and custom applications work with the same profile:

```bash
gosho login aws-dev
aws s3 ls --profile aws-dev   # works without 'aws sso login'
```

The profile name in gosho must match the `sso_session` name in `~/.aws/config`.

## Project structure

```
gosho/
├── main.go              # Entry point
├── Makefile             # Build/install (Linux/macOS)
├── build.ps1            # Build/install (Windows PowerShell)
├── cmd/
│   ├── login.go         # Interactive + preset + login all flow
│   ├── logout.go        # Clear token and credentials
│   ├── init.go          # Configure defaults
│   └── status.go        # Show cached profile status (with color)
├── sso/
│   ├── constant.go      # Regions, grant types, scopes
│   ├── auth.go          # OIDC device auth, browser launch, WSL detection
│   ├── token.go         # Token cache (per profile)
│   ├── account.go       # List accounts/roles, get credentials
│   └── credentials.go   # Write/remove ~/.aws/credentials
└── config/
    └── config.go        # YAML config (start URL, region, profile presets)
```

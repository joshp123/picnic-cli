# picnic CLI

CLI tool for managing a Picnic grocery cart.

## Install / Build

```bash
go build -o picnic
```

## Nix

```bash
# Build with Nix flakes
nix build
./result/bin/picnic-cli --help
```

## Clawdbot Plugin

This repo exports a `clawdbotPlugin` flake output for nix-clawdbot. nix-clawdbot
symlinks skills into `~/.clawdbot/skills/<plugin>/<skill>` and adds the plugin
packages to `PATH`, so no `skillsLoad.extraDirs` is needed.

## Usage

```bash
# Search products
picnic search "melk"

# Add to cart
picnic add <product_id> [count]

# Remove from cart
picnic remove <product_id> [count]

# View cart
picnic cart

# Clear cart
picnic clear

# List delivery slots
picnic slots

# Select a delivery slot
picnic slot set <slot_id>

# Start checkout
picnic checkout start

# Initiate payment
picnic checkout pay <order_id>

# Analyze purchase history
picnic analyze-orders
```

## Authentication

Provide credentials via environment variables:

- `PICNIC_EMAIL` (optional if `PICNIC_AUTH_FILE` is set)
- `PICNIC_PASSWORD` (optional if `PICNIC_AUTH_FILE` is set)
- `PICNIC_COUNTRY` (optional, default `NL`)
- `PICNIC_AUTH_FILE` (optional, path to credentials file)
- `PICNIC_TOKEN_FILE` (optional, default `~/.picnic-token`)

Auth tokens are cached at `PICNIC_TOKEN_FILE`.

Credentials file format (any of these):

```text
email=you@example.com
password=your-password
```

```text
[username]
you@example.com
[password]
your-password
```

## Data Files

The analyzer writes:

- `~/.picnic-history.json`
- `~/.picnic-preferences.json`

## License

MIT

---
name: picnic
description: Add groceries to Picnic shopping cart via voice/text commands.
homepage: https://picnic.app
metadata: {"clawdis":{"emoji":"cart","requires":{"env":["PICNIC_AUTH_FILE"]},"primaryEnv":"PICNIC_AUTH_FILE"}}
---

# Picnic Grocery Skill

Manage your Picnic grocery shopping cart via CLI.

## Setup

1. Set `PICNIC_EMAIL` and `PICNIC_PASSWORD` in config
2. Ensure `picnic` is on PATH (Clawdis plugin package)

## Commands

### Search for products
```bash
picnic search "melk"
```

### Add product to cart
```bash
picnic add PRODUCT_ID [count]
```

### View cart
```bash
picnic cart
```

### Remove from cart
```bash
picnic remove PRODUCT_ID [count]
```

### Clear cart
```bash
picnic clear
```

### Analyze order history
```bash
picnic analyze-orders
```

### List delivery slots
```bash
picnic slots
```

### Select delivery slot
```bash
picnic slot set SLOT_ID
```

### Start checkout
```bash
picnic checkout start
```

### Initiate payment
```bash
picnic checkout pay ORDER_ID
```

## Workflow

1. Search for product -> get product IDs
2. Add desired product by ID
3. View cart to confirm

## Notes

- Country: NL (Netherlands)
- Credentials stored in clawdis config
- Auth token cached in `~/.picnic-token`

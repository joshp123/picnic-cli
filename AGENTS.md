# AGENTS.md - picnic-cli

This file describes the Clawdis plugin knobs and how agents should obtain the
values. It is written for automation agents, not end users.

## Plugin knobs (values)

### PICNIC_AUTH_FILE (env value)
- **What it is**: Path to a file containing Picnic login credentials.
- **Format** (either style):
  ```
  email=USER_EMAIL
  password=USER_PASSWORD
  ```
  ```
  [username]
  USER_EMAIL
  [password]
  USER_PASSWORD
  ```
- **How to obtain**: Ask the operator for their Picnic login details, store in
  the secrets system (e.g. agenix), and point `PICNIC_AUTH_FILE` to the
  decrypted file path (example: `/run/agenix/picnic-auth`).
- **Why it matters**: Required for login.

### PICNIC_EMAIL / PICNIC_PASSWORD (optional env values)
- **What they are**: Picnic account credentials as direct env vars.
- **How to obtain**: Only use if explicitly requested; prefer `PICNIC_AUTH_FILE`.

### PICNIC_COUNTRY (optional env value)
- **What it is**: Country code for the Picnic storefront (default `NL`).
- **How to obtain**: Ask the operator which country their Picnic account is
  registered in.

### PICNIC_AUTH_FILE (optional env value)
- **What it is**: Path to the cached auth token file.
- **Default**: `~/.picnic-auth`
- **How to obtain**: Only override if the operator requests a custom path.

## Validation / smoke checks
- `picnic search "milch"`
- `picnic cart`

## Notes
- Auth tokens are cached at `PICNIC_TOKEN_FILE` (default `~/.picnic-token`) and
  can be deleted to force a re-login.
- Avoid hardcoding real credentials in repo files; use placeholders only.

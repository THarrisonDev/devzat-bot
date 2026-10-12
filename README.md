# devzat-bot

A [Devzat](https://github.com/quackduck/devzat) plugin (Go) with two quality-of-life
features for a team server:

- **Welcome** — DMs anyone who joins the chat a short guide (customizable).
- **Remind** — `remind <when> <message>` pings you later, e.g.
  `remind 2h review the deploy`.

## Build

```bash
go build -o devzat-bot .
```

## Configure

| Env | Default | Purpose |
|---|---|---|
| `DEVZAT_ADDRESS` | — | Plugin-API address, e.g. `localhost:5556` (arg overrides) |
| `DEVZAT_TOKEN` | — | Plugin API token (required) |
| `WELCOME_ENABLED` | `true` | DM new joiners the guide |
| `REMIND_ENABLED` | `true` | Enable the `remind` command |
| `WELCOME_FILE` | — | Path to a file with custom welcome text (`{{name}}` is substituted) |

## Usage

```
remind 30m check the build
remind 2h review #292
remind 1d12h follow up with the vendor
```

Units: `s`, `m`, `h`, `d` (and combinations, e.g. `1d12h`). The reminder is posted
back in the room where you set it and mentions your name (so it bells you); if that
room was since cleaned up, it falls back to `#main`.

**Note:** reminders are held in memory, so they don't survive a bot restart.

## Install (systemd)

```bash
sudo DEVZAT_TOKEN=dvz@... ./install.sh
```

Uses its own plugin-API token (mint one per plugin with `grant` in chat — see
`devzat-deploy`'s README).

## Operate

```bash
systemctl status devzat-bot
journalctl -u devzat-bot -f
systemctl restart devzat-bot
```

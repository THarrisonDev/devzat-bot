#!/usr/bin/env bash
# Build and install the Devzat welcome + remind bot.
#
# Run as root from this directory:
#   sudo ./install.sh
#
# Pass config inline to avoid editing the env file afterwards:
#   sudo DEVZAT_TOKEN=dvz@... ./install.sh
set -euo pipefail

SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN=/usr/local/bin/devzat-bot
SVC=/etc/systemd/system/devzat-bot.service
ENV_FILE=/etc/devzat-bot.env

echo "==> Building devzat-bot"
( cd "$SRC_DIR" && go build -o "$BIN" . )
chmod 755 "$BIN"

echo "==> Installing systemd unit"
install -m 0644 "$SRC_DIR/devzat-bot.service" "$SVC"

if [ ! -f "$ENV_FILE" ]; then
    echo "==> Creating $ENV_FILE"
    install -m 0600 /dev/null "$ENV_FILE"
    cat >"$ENV_FILE" <<'EOF'
# Devzat welcome + remind bot
# DEVZAT_TOKEN: mint one per plugin with `grant` (see devzat-deploy README)
DEVZAT_TOKEN=
WELCOME_ENABLED=true
REMIND_ENABLED=true
# WELCOME_FILE=/opt/devzat-bot/welcome.md
EOF
else
    echo "==> Keeping existing $ENV_FILE"
fi

update_env() {
    local key="$1" val="${2:-}"
    [ -n "$val" ] || return 0
    if grep -q "^${key}=" "$ENV_FILE"; then
        sed -i "s|^${key}=.*|${key}=${val}|" "$ENV_FILE"
    else
        printf '%s=%s\n' "$key" "$val" >>"$ENV_FILE"
    fi
}
update_env DEVZAT_TOKEN "${DEVZAT_TOKEN:-}"
update_env WELCOME_ENABLED "${WELCOME_ENABLED:-}"
update_env REMIND_ENABLED "${REMIND_ENABLED:-}"
chmod 600 "$ENV_FILE"

if ! grep -q '^DEVZAT_TOKEN=.\+' "$ENV_FILE"; then
    echo
    echo "WARNING: DEVZAT_TOKEN is empty — the bot will fail to start."
fi

echo "==> Starting service"
systemctl daemon-reload
systemctl enable devzat-bot
systemctl restart devzat-bot

echo
systemctl --no-pager --full status devzat-bot || true

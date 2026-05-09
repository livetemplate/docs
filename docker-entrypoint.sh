#!/bin/sh
# docs-site container entrypoint: runs the recipes binary on a loopback
# port and tinkerdown on the public port. Tinkerdown auto-proxies
# embed-lvt blocks to the recipes binary via http://localhost:9091.
#
# If `site` exits, `tinkerdown` keeps running but live demos break — so
# we trap SIGTERM and forward it to both.
set -eu

RECIPES_PORT="${RECIPES_PORT:-9091}"
PORT="${PORT:-8080}"

echo "[entrypoint] starting site on :${RECIPES_PORT}"
RECIPES_PORT="${RECIPES_PORT}" /usr/local/bin/site &
SITE_PID=$!

# Forward SIGTERM/SIGINT to both processes so Fly's graceful shutdown
# doesn't leave orphan pids.
trap 'kill -TERM "$SITE_PID" 2>/dev/null || true' TERM INT

echo "[entrypoint] starting tinkerdown on :${PORT}"
exec /usr/local/bin/tinkerdown serve --host 0.0.0.0 --port "${PORT}" /site

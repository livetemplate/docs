#!/bin/sh
# docs-site container entrypoint: runs the recipes binary on a loopback
# port and tinkerdown on the public port. Tinkerdown auto-proxies
# embed-lvt blocks to the recipes binary via http://localhost:9091.
#
# Both processes run in the background so the trap stays installed
# (`exec` would replace the shell and kill the trap). If either child
# exits, the container exits so Fly's healthcheck/restart kicks in.
set -eu

RECIPES_PORT="${RECIPES_PORT:-9091}"
PORT="${PORT:-8080}"

cleanup() {
    [ -n "${SITE_PID:-}" ] && kill -TERM "$SITE_PID" 2>/dev/null || true
    [ -n "${TINKERDOWN_PID:-}" ] && kill -TERM "$TINKERDOWN_PID" 2>/dev/null || true
    wait 2>/dev/null || true
}
trap cleanup TERM INT

echo "[entrypoint] starting site on :${RECIPES_PORT}"
RECIPES_PORT="${RECIPES_PORT}" /usr/local/bin/site &
SITE_PID=$!

echo "[entrypoint] starting tinkerdown on :${PORT}"
/usr/local/bin/tinkerdown serve --host 0.0.0.0 --port "${PORT}" /site &
TINKERDOWN_PID=$!

# Poll for either to exit; either way the container should die so Fly
# restarts it (a stale site means embed-lvt fails silently; a stale
# tinkerdown means docs serve 502s). 5s polling is fine for shutdown
# latency on a docs site. `wait -n` would be cleaner but isn't portable
# to BusyBox ash.
while kill -0 "$SITE_PID" 2>/dev/null && kill -0 "$TINKERDOWN_PID" 2>/dev/null; do
    sleep 5
done

echo "[entrypoint] one of {site, tinkerdown} exited; shutting down"
cleanup
exit 1

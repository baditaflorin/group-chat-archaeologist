#!/bin/bash
set -euo pipefail

if [[ -n "${PORT:-}" ]]; then
  port="$PORT"
else
  port="$(node - <<'JS'
const net = require('node:net');
const server = net.createServer();
server.listen(0, '127.0.0.1', () => {
  console.log(server.address().port);
  server.close();
});
JS
)"
fi
tmp_dir="$(mktemp -d)"
log_file="$(mktemp)"
server_pid=""

cleanup() {
  if [[ -n "$server_pid" ]]; then
    kill "$server_pid" >/dev/null 2>&1 || true
  fi
  rm -rf "$tmp_dir" "$log_file"
}
trap cleanup EXIT

if [[ "${SMOKE_SKIP_BUILD:-0}" != "1" ]]; then
  npm --prefix web run build
fi
cp -R "$(pwd)/docs" "$tmp_dir/group-chat-archaeologist"
node scripts/static-server.mjs "$tmp_dir" "$port" >"$log_file" 2>&1 &
server_pid="$!"

for _ in {1..40}; do
  if curl -fsS "http://127.0.0.1:${port}/group-chat-archaeologist/" >/dev/null; then
    break
  fi
  sleep 0.25
done

curl -fsS "http://127.0.0.1:${port}/group-chat-archaeologist/data/v1/chat-archaeology.json" >/dev/null
PLAYWRIGHT_BASE_URL="http://127.0.0.1:${port}/group-chat-archaeologist/" npm --prefix web run test:e2e

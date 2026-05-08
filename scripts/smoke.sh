#!/usr/bin/env bash
set -euo pipefail

port="${PORT:-$(python3 - <<'PY'
import socket

with socket.socket() as sock:
    sock.bind(("127.0.0.1", 0))
    print(sock.getsockname()[1])
PY
)}"
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

npm --prefix web run build
cp -R "$(pwd)/docs" "$tmp_dir/group-chat-archaeologist"
python3 -m http.server "$port" --bind 127.0.0.1 --directory "$tmp_dir" >"$log_file" 2>&1 &
server_pid="$!"

for _ in {1..40}; do
  if curl -fsS "http://127.0.0.1:${port}/group-chat-archaeologist/" >/dev/null; then
    break
  fi
  sleep 0.25
done

curl -fsS "http://127.0.0.1:${port}/group-chat-archaeologist/data/v1/chat-archaeology.json" >/dev/null
PLAYWRIGHT_BASE_URL="http://127.0.0.1:${port}/group-chat-archaeologist/" npm --prefix web run test:e2e

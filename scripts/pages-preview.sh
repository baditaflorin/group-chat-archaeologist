#!/usr/bin/env bash
set -euo pipefail

port="${PORT:-4174}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

cp -R "$(pwd)/docs" "$tmp_dir/group-chat-archaeologist"
echo "Serving https://baditaflorin.github.io/group-chat-archaeologist/ equivalent at:"
echo "http://127.0.0.1:${port}/group-chat-archaeologist/"
python3 -m http.server "$port" --bind 127.0.0.1 --directory "$tmp_dir"

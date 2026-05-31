#!/usr/bin/env bash
set -euo pipefail

mode="${1:-separate}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_dir="$root/dist"
binary_name="nozomi-relay"

mkdir -p "$out_dir"

case "$mode" in
  separate)
    (cd "$root/backend" && go build -o "$out_dir/$binary_name" ./cmd/server)
    echo "Built API-only binary: $out_dir/$binary_name"
    echo "Run frontend separately with: cd Nozomi-Admin && npm run dev"
    ;;
  embedded)
    (cd "$root/Nozomi-Admin" && npm run build)
    rm -rf "$root/backend/internal/server/web/dist"
    mkdir -p "$root/backend/internal/server/web"
    cp -R "$root/Nozomi-Admin/dist" "$root/backend/internal/server/web/dist"
    (cd "$root/backend" && go build -tags embedweb -o "$out_dir/$binary_name-embedded" ./cmd/server)
    echo "Built embedded full-stack binary: $out_dir/$binary_name-embedded"
    echo "Run it with: $out_dir/$binary_name-embedded"
    ;;
  *)
    echo "Usage: $0 [separate|embedded]" >&2
    exit 2
    ;;
esac

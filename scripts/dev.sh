#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export ADDR=127.0.0.1:8095
export FORUM_URL=http://localhost:8095/community
export COMMUNITY_DB="$PWD/data/development.db"
export COMMUNITY_INSECURE_COOKIES=1
unset COMMUNITY_ADMIN_PASSWORD_FILE
printf 'Local development: http://localhost:8095\nStop with Ctrl+C. No Docker or Cloudflare.\n'
exec go run .

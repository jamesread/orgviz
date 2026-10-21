#!/usr/bin/env bash
# Confirms the throwaway orgviz loaded the fictional seed org files.
# Org files are read at process start (there is no write API), so `make up`
# copies seed/ into the tmp layout before launching the binary.
# Usage: seed.sh <backend-base-url>
set -euo pipefail

base="${1:?backend base url}"
api="$base/api/clientapi.ClientToFrontend"

rpc() {
	local body="$2"
	if [ -z "$body" ]; then
		body='{}'
	fi
	curl -fsS -X POST -H 'Content-Type: application/json' -d "$body" "$api/$1"
}

settings="$(rpc getClientInitialSettings '{}')"
echo "$settings" | jq -e '.charts | length >= 2' >/dev/null
echo "$settings" | jq -e '.charts[] | select(.title=="Example Company")' >/dev/null
echo "$settings" | jq -e '.charts[] | select(.title=="Second Company")' >/dev/null

chart="$(rpc getChart '{"chartId":"0"}')"
echo "$chart" | jq -e '.people | length >= 5' >/dev/null

echo "Seeded charts: $(echo "$settings" | jq -r '[.charts[].title] | join(", ")')"

#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"
SOURCE_DIR="${GO_AGENT_SOURCE_DIR:-$REPO_ROOT/experiments/0006-support-harbo-evaluation}"
if [[ "$SOURCE_DIR" != /* ]]; then
  SOURCE_DIR="$REPO_ROOT/$SOURCE_DIR"
fi
STATE_DIR="${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor"
OUTPUT_DIR="${STATE_DIR}/agents"
OUTPUT_PATH="${GO_AGENT_BINARY:-${OUTPUT_DIR}/go-agent}"

die() { printf 'error: %s\n' "$*" >&2; exit 1; }

command -v go >/dev/null 2>&1 || die "Go is required to build the Go harness"
[[ -d "$SOURCE_DIR" ]] || die "Go harness directory not found: $SOURCE_DIR"

GOOS_TARGET="${GOOS_TARGET:-linux}"
if [[ -n "${GOARCH_TARGET:-}" ]]; then
  GOARCH_VALUE="$GOARCH_TARGET"
else
  DOCKER_ARCH="$(docker info --format '{{.Architecture}}' 2>/dev/null || true)"
  case "$DOCKER_ARCH" in
    aarch64|arm64) GOARCH_VALUE=arm64 ;;
    x86_64|amd64) GOARCH_VALUE=amd64 ;;
    armv7l|armv7) GOARCH_VALUE=arm ;;
    *) GOARCH_VALUE="$(go env GOARCH)" ;;
  esac
fi

mkdir -p "$OUTPUT_DIR"
printf 'Building Go harness: GOOS=%s GOARCH=%s\n' "$GOOS_TARGET" "$GOARCH_VALUE"

# The CLI entrypoint does not use the optional Lark connector. Build a clean
# temporary package without the Lark files or its external dependency, while
# retaining the main package's runLark symbol for compilation.
BUILD_DIR="$(mktemp -d "${TMPDIR:-/tmp}/harbor-go-agent.XXXXXX")"
cleanup() { rm -rf "$BUILD_DIR"; }
trap cleanup EXIT

if [[ -f "$SOURCE_DIR/go.mod" ]]; then
  cp -R "$SOURCE_DIR"/. "$BUILD_DIR"/
else
  find "$SOURCE_DIR" -maxdepth 1 -name '*.go' \
    ! -name 'lark_connector.go' \
    -exec cp {} "$BUILD_DIR" \;
fi

if [[ -f "$SOURCE_DIR/lark_connector.go" && ! -f "$SOURCE_DIR/go.mod" ]]; then
cat > "$BUILD_DIR/lark_connector_stub.go" <<'GO'
package main

import "context"

type LarkMessage struct {
	ChatId      string
	TextMessage TextMessage
}

type TextMessage struct {
	Text string `json:"text"`
}

type LarkConnector struct {
	InMessages  chan LarkMessage
	OutMessages chan LarkMessage
}

func NewLarkConnector(LarkConnectorConfig) LarkConnector {
	return LarkConnector{
		InMessages:  make(chan LarkMessage),
		OutMessages: make(chan LarkMessage),
	}
}

func (LarkConnector) start(context.Context) error { return nil }
GO
fi
(
  cd "$BUILD_DIR"
  if [[ -f go.mod ]]; then
    CGO_ENABLED=0 GOOS="$GOOS_TARGET" GOARCH="$GOARCH_VALUE" \
      go build -trimpath -o "$OUTPUT_PATH" .
  else
    GO111MODULE=off CGO_ENABLED=0 GOOS="$GOOS_TARGET" GOARCH="$GOARCH_VALUE" \
      go build -trimpath -o "$OUTPUT_PATH" .
  fi
)
chmod 0755 "$OUTPUT_PATH"
printf 'Go harness ready: %s\n' "$OUTPUT_PATH"

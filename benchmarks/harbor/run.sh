#!/usr/bin/env bash

set -Eeuo pipefail

HARBOR_VERSION="${HARBOR_VERSION:-0.24.0}"
HARBOR_ENV_NAME="${HARBOR_ENV_NAME:-harbor-${HARBOR_VERSION}}"
HARBOR_TASK_DIR="${HARBOR_TASK_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/tasks/hello-world}"
HARBOR_AGENT="${HARBOR_AGENT:-mini-swe-agent}"
HARBOR_MODEL="${HARBOR_MODEL:-openai/gpt-4o-mini}"
HARBOR_JOBS_DIR="${HARBOR_JOBS_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/jobs}"
HARBOR_AGENT_TIMEOUT_MULTIPLIER="${HARBOR_AGENT_TIMEOUT_MULTIPLIER:-5}"
HARBOR_AGENT_SETUP_TIMEOUT_MULTIPLIER="${HARBOR_AGENT_SETUP_TIMEOUT_MULTIPLIER:-20}"
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"
GO_AGENT_BINARY="${GO_AGENT_BINARY:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/agents/go-agent}"

# mini-swe-agent requires provider/model form. Allow a bare model id for an
# OpenAI-compatible endpoint, but always pass canonical form to Harbor.
if [[ "$HARBOR_MODEL" != */* ]]; then
  HARBOR_MODEL="openai/${HARBOR_MODEL}"
fi

if [[ "$HARBOR_AGENT" == "go-agent" ]]; then
  "$SCRIPT_DIR/deploy.sh"
  export GO_AGENT_BINARY
  HARBOR_AGENT="benchmarks.harbor.go_agent:GoAgent"
fi

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

command -v conda >/dev/null 2>&1 || die "Conda is required"
command -v docker >/dev/null 2>&1 || die "Docker is required; run setup.sh first"

if ! docker info >/dev/null 2>&1; then
  die "Docker is installed but its daemon is not running"
fi

CONDA_BASE="$(conda info --base)"
CONDA_ENV_DIR="${CONDA_ENV_DIR:-${CONDA_BASE}/envs/${HARBOR_ENV_NAME}}"
HARBOR_BIN="${CONDA_ENV_DIR}/bin/harbor"
if [[ ! -x "$HARBOR_BIN" || ! -f "$HARBOR_TASK_DIR/task.toml" ]]; then
  printf 'Harbor is not installed yet; preparing its isolated environment...\n'
  SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
  "$SCRIPT_DIR/setup.sh"
fi
[[ -x "$HARBOR_BIN" ]] || die "Harbor installation did not complete"
[[ -f "$HARBOR_TASK_DIR/task.toml" ]] || die "hello-world task preparation did not complete"
export PYTHONPATH="$REPO_ROOT${PYTHONPATH:+:$PYTHONPATH}"

if [[ -z "${PROXY_API_KEY:-}" && -z "${OPENAI_API_KEY:-}" ]]; then
  if [[ ! -t 0 ]]; then
    die "set OPENAI_API_KEY, or set PROXY_API_KEY and PROXY_BASE_URL for a gateway"
  fi
  printf 'No model API key found.\n'
  printf 'Use an OpenAI-compatible gateway? [y/N] '
  read -r USE_PROXY
  if [[ "$USE_PROXY" =~ ^[Yy]$ ]]; then
    read -r -p 'Gateway base URL (for example https://gateway.example/v1): ' PROXY_BASE_URL
    read -r -s -p 'Gateway API key: ' PROXY_API_KEY
    printf '\n'
    [[ -n "$PROXY_BASE_URL" && -n "$PROXY_API_KEY" ]] || die "gateway URL and API key are required"
    export PROXY_BASE_URL PROXY_API_KEY
  else
    read -r -s -p 'OpenAI API key: ' OPENAI_API_KEY
    printf '\n'
    [[ -n "$OPENAI_API_KEY" ]] || die "OpenAI API key is required"
    export OPENAI_API_KEY
  fi
fi

mkdir -p "$HARBOR_JOBS_DIR"

printf 'Harbor %s | agent: %s | model: %s\n' "$HARBOR_VERSION" "$HARBOR_AGENT" "$HARBOR_MODEL"

EXTRA_AUTH_ARGS=()
if [[ -n "${PROXY_API_KEY:-}" ]]; then
  [[ -n "${PROXY_BASE_URL:-}" ]] || die "PROXY_BASE_URL is required when using PROXY_API_KEY"
  # Harbor 0.24.0 passes agent environment values with --agent-env. The
  # mini-swe-agent OpenAI-compatible client reads these standard variables.
  EXTRA_AUTH_ARGS=(
    --agent-env "OPENAI_API_KEY=${PROXY_API_KEY}"
    --agent-env "OPENAI_BASE_URL=${PROXY_BASE_URL}"
  )
fi

if [[ "$HARBOR_AGENT" == "benchmarks.harbor.go_agent:GoAgent" ]]; then
  [[ -n "${OPENAI_API_KEY:-}" || -n "${PROXY_API_KEY:-}" ]] || \
    die "set OPENAI_API_KEY or PROXY_API_KEY for go-agent"
  [[ -n "${OPENAI_URL:-}" || -n "${PROXY_BASE_URL:-}" ]] || \
    die "set OPENAI_URL or PROXY_BASE_URL for go-agent"
  if [[ -z "${PROXY_API_KEY:-}" && -n "${OPENAI_API_KEY:-}" ]]; then
    EXTRA_AUTH_ARGS+=(--agent-env "OPENAI_API_KEY=${OPENAI_API_KEY}")
  fi
  if [[ -z "${OPENAI_URL:-}" ]]; then
    OPENAI_URL="${PROXY_BASE_URL%/}/chat/completions"
  fi
  export OPENAI_URL
  EXTRA_AUTH_ARGS+=(--agent-env "OPENAI_URL=${OPENAI_URL}")
  EXTRA_AUTH_ARGS+=(--agent-env "OPENAI_MODEL=${HARBOR_MODEL#*/}")
fi

if ((${#EXTRA_AUTH_ARGS[@]})); then
  exec "$HARBOR_BIN" run \
    --path "$HARBOR_TASK_DIR" \
    --agent "$HARBOR_AGENT" \
    --model "$HARBOR_MODEL" \
    --jobs-dir "$HARBOR_JOBS_DIR" \
    --agent-timeout-multiplier "$HARBOR_AGENT_TIMEOUT_MULTIPLIER" \
    --agent-setup-timeout-multiplier "$HARBOR_AGENT_SETUP_TIMEOUT_MULTIPLIER" \
    "${EXTRA_AUTH_ARGS[@]}" \
    "$@"
fi

exec "$HARBOR_BIN" run \
  --path "$HARBOR_TASK_DIR" \
  --agent "$HARBOR_AGENT" \
  --model "$HARBOR_MODEL" \
  --jobs-dir "$HARBOR_JOBS_DIR" \
  --agent-timeout-multiplier "$HARBOR_AGENT_TIMEOUT_MULTIPLIER" \
  --agent-setup-timeout-multiplier "$HARBOR_AGENT_SETUP_TIMEOUT_MULTIPLIER" \
  "$@"

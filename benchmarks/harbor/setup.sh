#!/usr/bin/env bash

set -Eeuo pipefail

# Harbor v0.24.0 currently has dependencies that do not build on Python 3.15.
# Keep it in a dedicated Conda environment outside this repository.
HARBOR_VERSION="${HARBOR_VERSION:-0.24.0}"
HARBOR_ENV_NAME="${HARBOR_ENV_NAME:-harbor-${HARBOR_VERSION}}"
HARBOR_TASK_DIR="${HARBOR_TASK_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/tasks/hello-world}"
HARBOR_DATASET="${HARBOR_DATASET:-}"
HARBOR_DATASET_DIR="${HARBOR_DATASET_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/datasets/terminal-bench}"

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

command -v conda >/dev/null 2>&1 || die "Conda is required"
command -v docker >/dev/null 2>&1 || die "Docker is required by Harbor's default local sandbox"

CONDA_BASE="$(conda info --base)"
CONDA_ENV_DIR="${CONDA_ENV_DIR:-${CONDA_BASE}/envs/${HARBOR_ENV_NAME}}"

if [[ ! -x "${CONDA_ENV_DIR}/bin/python" ]]; then
  printf 'Creating isolated Conda environment at %s...\n' "$CONDA_ENV_DIR"
  conda create --yes --prefix "$CONDA_ENV_DIR" "python=3.14" pip
fi

PYTHON="${CONDA_ENV_DIR}/bin/python"
[[ -x "$PYTHON" ]] || die "Conda environment Python was not created: $PYTHON"

PYTHON_VERSION="$($PYTHON -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')"
[[ "$PYTHON_VERSION" == 3.14 ]] || die "Harbor environment must use Python 3.14, found $PYTHON_VERSION"

printf 'Installing Harbor %s into the isolated Conda environment...\n' "$HARBOR_VERSION"
INSTALLED_VERSION="$($CONDA_ENV_DIR/bin/harbor --version 2>/dev/null || true)"
if [[ "$INSTALLED_VERSION" != "$HARBOR_VERSION" ]]; then
  "$PYTHON" -m pip install --upgrade "harbor==${HARBOR_VERSION}"
fi
HARBOR_MINI_ADAPTER="$($PYTHON -c 'import harbor.agents.installed.mini_swe_agent as m; print(m.__file__)')"
if ! grep -Fq -- '--with mcp' "$HARBOR_MINI_ADAPTER"; then
  printf 'Applying Harbor %s mini-swe-agent dependency compatibility patch...\n' "$HARBOR_VERSION"
  perl -0pi -e 's/--with litellm --with orjson --with fastapi/--with litellm --with orjson --with fastapi --with mcp/g' "$HARBOR_MINI_ADAPTER"
fi
grep -Fq -- '--with mcp' "$HARBOR_MINI_ADAPTER" || die "failed to patch mini-swe-agent mcp dependency"
"$CONDA_ENV_DIR/bin/harbor" --version

# Harbor 0.24.0 installs mini-swe-agent during every trial. Since the task
# image below already contains it, make that install step a fast no-op.
if ! grep -Fq 'harbor-skip-preinstalled-mini-swe-agent' "$HARBOR_MINI_ADAPTER"; then
  perl -0pi -e 's|"set -euo pipefail; "\n|"set -euo pipefail; # harbor-skip-preinstalled-mini-swe-agent "\n                "if command -v mini-swe-agent >/dev/null 2>&1; then mini-swe-agent --help; exit 0; fi; "\n|' "$HARBOR_MINI_ADAPTER"
fi
grep -Fq 'harbor-skip-preinstalled-mini-swe-agent' "$HARBOR_MINI_ADAPTER" || die "failed to patch preinstalled agent handling"

if [[ ! -f "$HARBOR_TASK_DIR/task.toml" ]]; then
  mkdir -p "$(dirname -- "$HARBOR_TASK_DIR")"
  "$CONDA_ENV_DIR/bin/harbor" tasks download hello-world/hello-world \
    --output-dir "$(dirname -- "$HARBOR_TASK_DIR")" --export
fi
[[ -f "$HARBOR_TASK_DIR/task.toml" ]] || die "hello-world task was not downloaded"

# Tell Harbor to use the image built below. Without this explicit task-level
# reference, Harbor 0.24.0 rebuilds the Dockerfile for every trial.
if ! grep -Fq 'docker_image = "harbor-hello-world-mini-swe-agent"' "$HARBOR_TASK_DIR/task.toml"; then
  perl -0pi -e 's|\[environment\]\n|[environment]\ndocker_image = "harbor-hello-world-mini-swe-agent"\n|' "$HARBOR_TASK_DIR/task.toml"
fi
grep -Fq 'docker_image = "harbor-hello-world-mini-swe-agent"' "$HARBOR_TASK_DIR/task.toml" || die "failed to configure prebuilt hello-world image"

# Build the agent into the task image so each trial does not reinstall it over
# the network. This file lives under the user state directory, never in git.
TASK_DOCKERFILE="$HARBOR_TASK_DIR/environment/Dockerfile"
if ! grep -Fq 'harbor-preinstalled-mini-swe-agent' "$TASK_DOCKERFILE"; then
  cat >> "$TASK_DOCKERFILE" <<'DOCKERFILE'

# harbor-preinstalled-mini-swe-agent
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates git build-essential \
    && rm -rf /var/lib/apt/lists/* \
    && curl -LsSf https://astral.sh/uv/install.sh | sh \
    && /root/.local/bin/uv python install 3.12 \
    && /root/.local/bin/uv tool install --python 3.12 mini-swe-agent --with litellm --with orjson --with fastapi --with mcp
ENV PATH="/root/.local/bin:$PATH"
DOCKERFILE
fi

# The registry hello-world verifier currently installs curl, uv, pytest, and
# its CTRF plugin on every trial. Preinstall the verifier toolchain in the
# reusable image so verification does not depend on apt/PyPI availability.
if ! grep -Fq 'harbor-preinstalled-verifier' "$TASK_DOCKERFILE"; then
  cat >> "$TASK_DOCKERFILE" <<'DOCKERFILE'

# harbor-preinstalled-verifier
RUN /root/.local/bin/uv tool install --python 3.12 \
      --with pytest-json-ctrf==0.3.5 \
      pytest==8.4.1
DOCKERFILE
fi

TASK_TEST_SCRIPT="$HARBOR_TASK_DIR/tests/test.sh"
if ! grep -Fq 'harbor-preinstalled-verifier' "$TASK_TEST_SCRIPT"; then
  perl -0pi -e 's|^#!/bin/bash\n.*?^if \[ \$\? -eq 0 \]; then\n.*?^fi\n|#!/bin/bash\n\n# harbor-preinstalled-verifier\nset -e\n\npytest --ctrf /logs/verifier/ctrf.json /tests/test_state.py -rA\necho 1 > /logs/verifier/reward.txt\n|ms' "$TASK_TEST_SCRIPT"
fi
grep -Fq 'harbor-preinstalled-verifier' "$TASK_TEST_SCRIPT" || die "failed to make verifier offline"

# Download the selected benchmark outside the repository. Dataset task images
# are built by Harbor on demand; the task sources themselves are cached here
# so later runs do not download the benchmark again.
if [[ -n "$HARBOR_DATASET" && "$HARBOR_DATASET" == terminal-bench/terminal-bench@* ]]; then
  DATASET_NAME_DIR="$HARBOR_DATASET_DIR/terminal-bench"
elif [[ -n "$HARBOR_DATASET" ]]; then
  DATASET_SLUG="${HARBOR_DATASET%@*}"
  DATASET_SLUG="${DATASET_SLUG##*/}"
  DATASET_NAME_DIR="$HARBOR_DATASET_DIR/$DATASET_SLUG"
fi
if [[ ! -f "$DATASET_NAME_DIR/dataset.toml" && ! -d "$DATASET_NAME_DIR" ]]; then
  mkdir -p "$HARBOR_DATASET_DIR"
  printf 'Downloading Harbor dataset %s...\n' "$HARBOR_DATASET"
  "$CONDA_ENV_DIR/bin/harbor" datasets download "$HARBOR_DATASET" \
    --output-dir "$HARBOR_DATASET_DIR" --export
fi
if [[ -n "$HARBOR_DATASET" ]]; then
  [[ -d "$DATASET_NAME_DIR" ]] || die "dataset was not downloaded: $DATASET_NAME_DIR"
fi

# Harbor's v0.24.0 adapter otherwise reinstalls the agent even when the image
# already contains it. Make setup a no-op when the preinstalled CLI exists.
printf 'Building the reusable hello-world Docker image with mini-swe-agent...\n'
docker build --tag harbor-hello-world-mini-swe-agent "$HARBOR_TASK_DIR/environment"

printf '\nHarbor is ready. Environment: %s\nTask image prepared at: %s\nRun the hello-world trial with:\n' "$CONDA_ENV_DIR" "$HARBOR_TASK_DIR"
printf '  %s/benchmarks/harbor/run.sh\n' "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

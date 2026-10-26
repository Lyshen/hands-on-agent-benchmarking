#!/usr/bin/env bash

set -Eeuo pipefail

# Harbor v0.24.0 currently has dependencies that do not build on Python 3.15.
# Keep it in a dedicated Conda environment outside this repository.
HARBOR_VERSION="${HARBOR_VERSION:-0.24.0}"
HARBOR_ENV_NAME="${HARBOR_ENV_NAME:-harbor-${HARBOR_VERSION}}"
HARBOR_TASK_DIR="${HARBOR_TASK_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/tasks/hello-world}"
HARBOR_DATASET="${HARBOR_DATASET:-}"
HARBOR_DATASET_DIR="${HARBOR_DATASET_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/datasets/terminal-bench}"
HARBOR_MINI_RUNTIME_DIR="${HARBOR_MINI_RUNTIME_DIR:-${XDG_STATE_HOME:-${HOME}/.local/state}/hands-on-agent-benchmarking/harbor/runtimes/mini-swe-agent-linux-amd64}"
HARBOR_MINI_IMAGE="${HARBOR_MINI_IMAGE:-harbor-terminal-bench-mini-swe-agent-amd64}"

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
  perl -0pi -e 's|"set -euo pipefail; "\n|"set -euo pipefail; # harbor-skip-preinstalled-mini-swe-agent "\n                "export PATH=\"$HOME/.local/bin:$PATH\"; "\n                "if command -v mini-swe-agent >/dev/null 2>&1; then mini-swe-agent --help; exit 0; fi; "\n|' "$HARBOR_MINI_ADAPTER"
fi
# Older patched copies may contain the marker but not the PATH prelude. Keep
# the check self-contained because Harbor does not necessarily pass agent-env
# values to the setup phase.
if ! grep -Fq 'export PATH="$HOME/.local/bin:$PATH"' "$HARBOR_MINI_ADAPTER"; then
  perl -0pi -e 's|("set -euo pipefail; # harbor-skip-preinstalled-mini-swe-agent "\n)|$1                "export PATH=\\"\$HOME/.local/bin:\$PATH\\"; "\n                "if command -v mini-swe-agent >/dev/null 2>&1; then mini-swe-agent --help; exit 0; fi; "\n|' "$HARBOR_MINI_ADAPTER"
fi
# Skip Harbor's apt-get setup when run.sh provides a validated preinstalled
# mini-swe-agent runtime. This probe must happen before
# ensure_system_dependencies(), otherwise every task pays the apt cost.
if ! grep -Fq 'harbor-probe-before-system-dependencies' "$HARBOR_MINI_ADAPTER" || grep -Fq 'export PATH="$HOME/.local/bin:$PATH"' "$HARBOR_MINI_ADAPTER"; then
  perl -0pi -e 's|    async def install\(self, environment: BaseEnvironment\) -> None:\n|    async def install(self, environment: BaseEnvironment) -> None:\n        # harbor-probe-before-system-dependencies\n        preinstalled = await self.exec_as_agent(environment, command=(\n            '\''export PATH="/root/.local/python/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"; '\''\n            '\''export PYTHONPATH="/root/.local/lib/python3.12/site-packages"; '\''\n            '\''command -v mini-swe-agent >/dev/null 2>&1 && mini-swe-agent --help >/dev/null'\''\n        ))\n        if preinstalled.return_code == 0:\n            return\n|' "$HARBOR_MINI_ADAPTER"
fi
# Repair an earlier version of this compatibility patch that expanded
# $HOME/$PATH while generating the Python source (it produced /.local/bin).
perl -0pi -e 's|export PATH="/\\.local/bin:";|export PATH="/root/.local/python/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin";|g' "$HARBOR_MINI_ADAPTER"
grep -Fq 'harbor-probe-before-system-dependencies' "$HARBOR_MINI_ADAPTER" || die "failed to patch mini-swe-agent preinstall probe"
grep -Fq 'harbor-skip-preinstalled-mini-swe-agent' "$HARBOR_MINI_ADAPTER" || die "failed to patch preinstalled agent handling"
grep -Fq 'export PATH=' "$HARBOR_MINI_ADAPTER" || die "failed to patch mini-swe-agent PATH handling"

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
DATASET_NAME_DIR=""
if [[ -n "$HARBOR_DATASET" && "$HARBOR_DATASET" == terminal-bench/terminal-bench@* ]]; then
  DATASET_NAME_DIR="$HARBOR_DATASET_DIR/terminal-bench"
elif [[ -n "$HARBOR_DATASET" ]]; then
  DATASET_SLUG="${HARBOR_DATASET%@*}"
  DATASET_SLUG="${DATASET_SLUG##*/}"
  DATASET_NAME_DIR="$HARBOR_DATASET_DIR/$DATASET_SLUG"
fi
if [[ -n "$HARBOR_DATASET" && ! -d "$DATASET_NAME_DIR" ]]; then
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

# Terminal-Bench's published task images are linux/amd64 even on an Apple
# Silicon host. Build a matching image once, then export its runtime. The
# runtime is mounted into official task images by run.sh instead of
# reinstalling mini-swe-agent for every trial.
if ! docker image inspect "$HARBOR_MINI_IMAGE" >/dev/null 2>&1; then
  die "AMD64 mini-swe-agent image is missing: $HARBOR_MINI_IMAGE"
fi
if [[ ! -e "$HARBOR_MINI_RUNTIME_DIR/bin/mini-swe-agent" || ! -x "$HARBOR_MINI_RUNTIME_DIR/python/bin/python3.12" ]]; then
  RUNTIME_TMP="$(mktemp -d "${TMPDIR:-/tmp}/harbor-mini-runtime.XXXXXX")"
  RUNTIME_CONTAINER="$(docker create --platform linux/amd64 "$HARBOR_MINI_IMAGE")"
  cleanup_runtime() {
    docker rm "$RUNTIME_CONTAINER" >/dev/null 2>&1 || true
    rm -rf "$RUNTIME_TMP"
  }
  trap cleanup_runtime EXIT
  mkdir -p "$RUNTIME_TMP"
  # Copy the complete .local tree, then rewrite container-absolute uv links
  # below. The resulting tree is mounted at the same path in task containers.
  docker cp "$RUNTIME_CONTAINER:/root/.local/." "$RUNTIME_TMP/"
  # The console script generated by uv has a /usr/local/bin/python shebang.
  # That path exists in the builder image but not in every Terminal-Bench
  # task image. Export the interpreter and stdlib alongside the packages so
  # the runtime is independent of the task image's Python installation.
  mkdir -p "$RUNTIME_TMP/python/bin" "$RUNTIME_TMP/python/lib"
  docker cp "$RUNTIME_CONTAINER:/usr/local/bin/python3.12" "$RUNTIME_TMP/python/bin/python3.12"
  docker cp "$RUNTIME_CONTAINER:/usr/local/lib/python3.12" "$RUNTIME_TMP/python/lib/"
  docker cp "$RUNTIME_CONTAINER:/usr/local/lib/libpython3.12.so.1.0" "$RUNTIME_TMP/python/lib/libpython3.12.so.1.0"
  perl -0pi -e 's|\A#!.*\n|#!/usr/bin/env python3.12\n|' "$RUNTIME_TMP/bin/mini-swe-agent"
  # Make the exported tree relocatable: /root/.local is the mount point in
  # every task container, so links must resolve inside the mounted tree.
  find "$RUNTIME_TMP" -type l -print0 | while IFS= read -r -d '' link; do
    target="$(readlink "$link")"
    case "$target" in
      /root/.local/*)
        relative="$(python3 -c 'import os,sys; print(os.path.relpath(sys.argv[1], os.path.dirname(sys.argv[2])))' \
          "${RUNTIME_TMP}${target#/root/.local}" "$link")"
        rm "$link"
        ln -s "$relative" "$link"
        ;;
    esac
  done
  mkdir -p "$(dirname -- "$HARBOR_MINI_RUNTIME_DIR")"
  mv "$RUNTIME_TMP" "$HARBOR_MINI_RUNTIME_DIR"
  docker rm "$RUNTIME_CONTAINER" >/dev/null
  trap - EXIT
fi
[[ -e "$HARBOR_MINI_RUNTIME_DIR/bin/mini-swe-agent" ]] || die "mini-swe runtime export failed"

printf '\nHarbor is ready. Environment: %s\nTask image prepared at: %s\nRun the hello-world trial with:\n' "$CONDA_ENV_DIR" "$HARBOR_TASK_DIR"
printf '  %s/benchmarks/harbor/run.sh\n' "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

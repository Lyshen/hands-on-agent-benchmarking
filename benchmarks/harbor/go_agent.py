"""Harbor adapter for the 0006 Go harness."""

from __future__ import annotations

import os
import shlex
from pathlib import Path
from typing import override

from harbor.agents.capabilities import AgentCapabilities
from harbor.agents.installed.base import BaseInstalledAgent
from harbor.agents.model_connection import ModelConnectionSpec
from harbor.environments.base import BaseEnvironment
from harbor.models.agent.context import AgentContext


_STATE_DIR = Path(
    os.environ.get("XDG_STATE_HOME", str(Path.home() / ".local" / "state"))
) / "hands-on-agent-benchmarking" / "harbor"
_DEFAULT_BINARY = _STATE_DIR / "agents" / "go-agent"
_REMOTE_BINARY = "/installed-agent/go-agent"


class GoAgent(BaseInstalledAgent):
    """Upload and run the locally-built 0006 Go CLI in the trial container."""

    capabilities = AgentCapabilities()
    MODEL_CONNECTION = ModelConnectionSpec(
        default_provider="openai",
        api_key_envs=("OPENAI_API_KEY",),
        base_url_envs=("OPENAI_URL",),
    )

    @staticmethod
    @override
    def name() -> str:
        return "go-agent"

    @override
    def version(self) -> str | None:
        return None

    @override
    async def install(self, environment: BaseEnvironment) -> None:
        source = Path(os.environ.get("GO_AGENT_BINARY", str(_DEFAULT_BINARY)))
        if not source.is_file():
            raise FileNotFoundError(
                f"Go harness binary not found: {source}. "
                "Run benchmarks/harbor/deploy.sh first."
            )
        await environment.upload_file(source, _REMOTE_BINARY)
        await self.exec_as_root(
            environment, command=f"chmod 0755 {shlex.quote(_REMOTE_BINARY)}"
        )
        check = await environment.exec(command=f"{shlex.quote(_REMOTE_BINARY)} --help")
        if check.return_code != 0:
            raise RuntimeError(
                "Uploaded Go harness failed its --help check: "
                f"{check.stdout or ''}{check.stderr or ''}"
            )

    @override
    async def run(
        self,
        instruction: str,
        environment: BaseEnvironment,
        context: AgentContext,
    ) -> None:
        access = self.model_connection
        api_key = access.api_key or self._get_env("OPENAI_API_KEY")
        url = self._get_env("OPENAI_URL") or access.configured_base_url
        model = self._get_env("OPENAI_MODEL") or self.model_name
        if not api_key:
            raise ValueError("OPENAI_API_KEY is required for go-agent")
        if not url:
            raise ValueError("OPENAI_URL must be the full chat-completions URL")
        if not model:
            raise ValueError("OPENAI_MODEL or --model is required for go-agent")

        env = {
            "OPENAI_API_KEY": api_key,
            "OPENAI_URL": url,
            "OPENAI_MODEL": model.removeprefix("openai/"),
        }
        command = (
            f"{shlex.quote(_REMOTE_BINARY)} --mode cli --task {shlex.quote(instruction)} "
            "2>&1 | tee /logs/agent/go-agent.log"
        )
        result = await self.exec_as_agent(environment, command=command, env=env)
        if result.stdout:
            self.logger.info("Go harness output:\n%s", result.stdout)

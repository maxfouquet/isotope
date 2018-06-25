from typing import Any, Dict, List

import toml


class RunnerConfig:
    def __init__(self, topology_paths: List[str]) -> None:
        self.topology_paths = topology_paths


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])

    return RunnerConfig(topology_paths=topology_paths)


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)

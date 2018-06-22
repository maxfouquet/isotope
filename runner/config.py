from typing import Dict

import toml


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    return RunnerConfig(**d)

class RunnerConfig:
    def __init__(self, fortio_client_image: str) -> None:
        pass

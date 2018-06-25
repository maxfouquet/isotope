from typing import Any, Dict, List

import toml


class RunnerConfig:
    def __init__(self, topology_paths: List[str], istio_hub: str,
                 istio_tag: str) -> None:
        self.topology_paths = topology_paths
        self.istio_hub = istio_hub
        self.istio_tag = istio_tag


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])

    istio = d['istio']
    istio_hub = istio['hub']
    istio_tag = istio['tag']

    return RunnerConfig(
        topology_paths=topology_paths,
        istio_hub=istio_hub,
        istio_tag=istio_tag)


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)

"""Read runner configuration from a dict or TOML."""

from typing import Any, Dict, List, Optional

import toml


class RunnerConfig:
    """Represents the intermediary between a config file"""

    def __init__(self, topology_paths: List[str], environments: List[str],
                 istio_archive_url: str, server_image: str, client_image: str,
                 client_qps: Optional[int], client_duration: str,
                 client_num_conc_conns: int) -> None:
        self.topology_paths = topology_paths
        self.environments = environments
        self.istio_archive_url = istio_archive_url
        self.server_image = server_image
        self.client_image = client_image
        self.client_qps = client_qps
        self.client_duration = client_duration
        self.client_num_conc_conns = client_num_conc_conns

    def labels(self) -> Dict[str, str]:
        """Returns the static labels for Prometheus for this configuration."""
        return {
            'istio_archive_url': self.istio_archive_url,
            'server_image': self.server_image,
            'client_image': self.client_image,
            'client_qps': str(self.client_qps),
            'client_duration': self.client_duration,
            'client_num_concurrent_connections':
            str(self.client_num_conc_conns),
        }


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])
    environments = d.get('environments', [])

    istio = d['istio']
    istio_archive_url = istio['archive_url']

    server = d['server']
    server_image = server['image']

    client = d['client']
    client_image = client['image']
    client_qps = client['qps']
    if client_qps == 'max':
        client_qps = None
    else:
        # Must coerce into integer, otherwise not a valid QPS.
        client_qps = int(client_qps)
    client_duration = client['duration']
    client_num_conc_conns = client['num_concurrent_connections']

    return RunnerConfig(
        topology_paths=topology_paths,
        environments=environments,
        istio_archive_url=istio_archive_url,
        server_image=server_image,
        client_image=client_image,
        client_qps=client_qps,
        client_duration=client_duration,
        client_num_conc_conns=client_num_conc_conns)


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)

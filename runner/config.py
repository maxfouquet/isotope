import enum
from typing import Any, Dict, Iterable, List, Tuple

import toml


class Environment(enum.Enum):
    NONE = 0
    ISTIO = 1


class RunnerConfig:
    def __init__(self, topology_paths: List[str],
                 environments: List[Environment], istio_hub: str,
                 istio_tag: str, istio_build: bool, cluster_name: str,
                 cluster_zone: str, cluster_version: str, cluster_create: bool,
                 server_machine_type: str, server_disk_size_gb: int,
                 server_num_nodes: int, server_image: str,
                 client_machine_type: str, client_disk_size_gb: int,
                 client_image: str, client_args: List[str]) -> None:
        self.topology_paths = topology_paths
        self.environments = environments
        self.istio_hub = istio_hub
        self.istio_tag = istio_tag
        self.should_build_istio = istio_build
        self.cluster_name = cluster_name
        self.cluster_zone = cluster_zone
        self.cluster_version = cluster_version
        self.should_create_cluster = cluster_create
        self.server_machine_type = server_machine_type
        self.server_disk_size_gb = server_disk_size_gb
        self.server_num_nodes = server_num_nodes
        self.server_image = server_image
        self.client_machine_type = client_machine_type
        self.client_disk_size_gb = client_disk_size_gb
        self.client_image = client_image
        self.client_args = client_args

    def labels(self) -> Dict[str, str]:
        return {
            'istio_hub': self.istio_hub,
            'istio_tag': self.istio_tag,
            'cluster_version': self.cluster_version,
            'cluster_zone': self.cluster_zone,
            'server_machine_type': self.server_machine_type,
            'server_disk_size_gb': str(self.server_disk_size_gb),
            'server_num_nodes': str(self.server_num_nodes),
            'server_image': self.server_image,
            'client_machine_type': self.client_machine_type,
            'client_disk_size_gb': str(self.client_disk_size_gb),
            'client_image': self.client_image,
        }


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])
    environment_strs = d.get('environments', [])
    environments = [Environment[s] for s in environment_strs]

    istio = d['istio']
    istio_hub = istio['hub']
    istio_tag = istio['tag']
    istio_build = istio['build']

    cluster = d['cluster']
    cluster_name = cluster['name']
    cluster_zone = cluster['zone']
    cluster_version = cluster['version']
    cluster_create = cluster['create']

    server = d['server']
    server_machine_type = server['machine_type']
    server_disk_size_gb = server['disk_size_gb']
    server_num_nodes = server['num_nodes']
    server_image = server['image']

    client = d['client']
    client_machine_type = client['machine_type']
    client_disk_size_gb = client['disk_size_gb']
    client_image = client['image']
    client_args = client['args']

    return RunnerConfig(
        topology_paths=topology_paths,
        environments=environments,
        istio_hub=istio_hub,
        istio_tag=istio_tag,
        istio_build=istio_build,
        cluster_name=cluster_name,
        cluster_zone=cluster_zone,
        cluster_version=cluster_version,
        cluster_create=cluster_create,
        server_machine_type=server_machine_type,
        server_disk_size_gb=server_disk_size_gb,
        server_image=server_image,
        server_num_nodes=server_num_nodes,
        client_machine_type=client_machine_type,
        client_disk_size_gb=client_disk_size_gb,
        client_image=client_image,
        client_args=client_args)


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)

from typing import Any, Dict, List

import toml


class RunnerConfig:
    def __init__(self, topology_paths: List[str], istio_hub: str,
                 istio_tag: str, cluster_name: str, cluster_zone: str,
                 cluster_version: str, server_machine_type: str,
                 server_disk_size_gb: int, server_num_nodes: int,
                 server_image: str, client_machine_type: str,
                 client_disk_size_gb: int, client_image: str,
                 client_args: List[str]) -> None:
        self.topology_paths = topology_paths
        self.istio_hub = istio_hub
        self.istio_tag = istio_tag
        self.cluster_name = cluster_name
        self.cluster_zone = cluster_zone
        self.cluster_version = cluster_version
        self.server_machine_type = server_machine_type
        self.server_disk_size_gb = server_disk_size_gb
        self.server_num_nodes = server_num_nodes
        self.server_image = server_image
        self.client_machine_type = client_machine_type
        self.client_disk_size_gb = client_disk_size_gb
        self.client_image = client_image
        self.client_args = client_args


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])

    istio = d['istio']
    istio_hub = istio['hub']
    istio_tag = istio['tag']

    cluster = d['cluster']
    cluster_name = cluster['name']
    cluster_zone = cluster['zone']
    cluster_version = cluster['version']

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
        istio_hub=istio_hub,
        istio_tag=istio_tag,
        cluster_name=cluster_name,
        cluster_zone=cluster_zone,
        cluster_version=cluster_version,
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

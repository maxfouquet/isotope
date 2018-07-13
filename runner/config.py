from typing import Any, Dict, List, Optional

import toml


class RunnerConfig:
    """Represents the intermediary between a config file"""

    def __init__(self, topology_paths: List[str], environments: List[str],
                 should_tear_down: bool, should_tear_down_on_error: bool,
                 istio_hub: str, istio_tag: str, istio_build: bool,
                 cluster_project_id: str, cluster_name: str, cluster_zone: str,
                 cluster_version: str, cluster_create: bool,
                 server_machine_type: str, server_disk_size_gb: int,
                 server_num_nodes: int, server_image: str,
                 client_machine_type: str, client_disk_size_gb: int,
                 client_image: str, client_qps: Optional[int],
                 client_duration: str, client_num_conc_conns: int) -> None:
        self.topology_paths = topology_paths
        self.environments = environments
        self.should_tear_down = should_tear_down
        self.should_tear_down_on_error = should_tear_down_on_error
        self.istio_hub = istio_hub
        self.istio_tag = istio_tag
        self.should_build_istio = istio_build
        self.cluster_project_id = cluster_project_id
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
        self.client_qps = client_qps
        self.client_duration = client_duration
        self.client_num_conc_conns = client_num_conc_conns

    def labels(self) -> Dict[str, str]:
        """Returns the static labels for Prometheus for this configuration."""
        return {
            # Disabled until addressing Stackdriver's 10-label limit.
            # 'istio_hub': self.istio_hub,
            # 'istio_tag': self.istio_tag,
            # 'cluster_version': self.cluster_version,
            # 'cluster_zone': self.cluster_zone,
            # 'server_machine_type': self.server_machine_type,
            # 'server_disk_size_gb': str(self.server_disk_size_gb),
            # 'server_num_nodes': str(self.server_num_nodes),
            # 'server_image': self.server_image,
            # 'client_machine_type': self.client_machine_type,
            # 'client_disk_size_gb': str(self.client_disk_size_gb),
            # 'client_image': self.client_image,
            # 'client_qps': str(self.client_qps),
            # 'client_duration': self.client_duration,
            # 'client_num_concurrent_connections': str(self.client_num_conc_conns),
        }


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])
    environments = d.get('environments', [])

    should_tear_down = d.get('tear_down', True)
    should_tear_down_on_error = d.get('tear_down_on_error', True)

    istio = d['istio']
    istio_hub = istio['hub']
    istio_tag = istio['tag']
    istio_build = istio['build']

    cluster = d['cluster']
    cluster_project_id = cluster['project_id']
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
        should_tear_down=should_tear_down,
        should_tear_down_on_error=should_tear_down_on_error,
        istio_hub=istio_hub,
        istio_tag=istio_tag,
        istio_build=istio_build,
        cluster_project_id=cluster_project_id,
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
        client_qps=client_qps,
        client_duration=client_duration,
        client_num_conc_conns=client_num_conc_conns)


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)

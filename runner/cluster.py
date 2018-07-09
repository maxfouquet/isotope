import logging
import os

from . import consts, resources, sh, wait


def setup(name: str, zone: str, version: str, service_graph_machine_type: str,
          service_graph_disk_size_gb: int, service_graph_num_nodes: int,
          client_machine_type: str, client_disk_size_gb: int) -> None:
    """Creates and sets up a GKE cluster.

    Args:
        name: name of the GKE cluster
        zone: GCE zone (e.g. "us-central1-a")
        version: GKE version (e.g. "1.9.7-gke.3")
        service_graph_machine_type: GCE type of service machines
        service_graph_disk_size_gb: disk size of service machines in gigabytes
        service_graph_num_nodes: number of machines in the service graph pool
        client_machine_type: GCE type of client machine
        client_disk_size_gb: disk size of client machine in gigabytes
    """
    _create_cluster(name, zone, version, 'n1-standard-1', 16, 1)
    _create_cluster_role_binding()

    _create_persistent_volume()
    _initialize_helm()
    _helm_add_prometheus_operator()
    _helm_add_prometheus()

    _create_service_graph_node_pool(service_graph_num_nodes,
                                    service_graph_machine_type,
                                    service_graph_disk_size_gb)
    _create_client_node_pool(client_machine_type, client_disk_size_gb)


def _create_cluster(name: str, zone: str, version: str, machine_type: str,
                    disk_size_gb: int, num_nodes: int) -> None:
    logging.info('creating cluster "%s"', name)
    sh.run_gcloud(
        [
            'container', 'clusters', 'create', name, '--zone', zone,
            '--cluster-version', version, '--machine-type', machine_type,
            '--disk-size',
            str(disk_size_gb), '--num-nodes',
            str(num_nodes)
        ],
        check=True)
    sh.run_gcloud(['config', 'set', 'container/cluster', name], check=True)
    sh.run_gcloud(
        ['container', 'clusters', 'get-credentials', name], check=True)


def _create_service_graph_node_pool(num_nodes: int, machine_type: str,
                                    disk_size_gb: int) -> None:
    logging.info('creating service graph node-pool')
    _create_node_pool(consts.SERVICE_GRAPH_NODE_POOL_NAME, num_nodes,
                      machine_type, disk_size_gb)


def _create_client_node_pool(machine_type: str, disk_size_gb: int) -> None:
    logging.info('creating client node-pool')
    _create_node_pool(consts.CLIENT_NODE_POOL_NAME, 1, machine_type,
                      disk_size_gb)


def _create_node_pool(name: str, num_nodes: int, machine_type: str,
                      disk_size_gb: int) -> None:
    sh.run_gcloud(
        [
            'container', 'node-pools', 'create', name, '--machine-type',
            machine_type, '--num-nodes',
            str(num_nodes), '--disk-size',
            str(disk_size_gb)
        ],
        check=True)


def _create_cluster_role_binding() -> None:
    logging.info('creating cluster-admin-binding')
    proc = sh.run_gcloud(['config', 'get-value', 'account'], check=True)
    account = proc.stdout
    sh.run_kubectl(
        [
            'create', 'clusterrolebinding', 'cluster-admin-binding',
            '--clusterrole', 'cluster-admin', '--user', account
        ],
        check=True)


def _create_persistent_volume() -> None:
    logging.info('creating persistent volume')
    sh.run_kubectl(
        ['apply', '-f', resources.PERSISTENT_VOLUME_YAML_PATH], check=True)


def _initialize_helm() -> None:
    logging.info('initializing Helm')
    sh.run_kubectl(
        ['create', '-f', resources.HELM_SERVICE_ACCOUNT_YAML_PATH], check=True)
    sh.run_helm(['init', '--service-account', 'tiller', '--wait'], check=True)
    sh.run_helm(
        [
            'repo', 'add', 'coreos',
            'https://s3-eu-west-1.amazonaws.com/coreos-charts/stable'
        ],
        check=True)


def _helm_add_prometheus_operator() -> None:
    logging.info('installing coreos/prometheus-operator')
    sh.run_helm(
        [
            'install', 'coreos/prometheus-operator', '--name',
            'prometheus-operator', '--namespace', consts.MONITORING_NAMESPACE
        ],
        check=True)


def _helm_add_prometheus() -> None:
    logging.info('installing coreos/prometheus')
    sh.run_helm(
        [
            'install', 'coreos/prometheus', '--name', 'prometheus',
            '--namespace', consts.MONITORING_NAMESPACE, '--values',
            resources.PROMETHEUS_STORAGE_VALUES_YAML_PATH
        ],
        check=True)
    wait.until_stateful_sets_are_ready(consts.MONITORING_NAMESPACE)

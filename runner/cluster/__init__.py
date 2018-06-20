import logging
import os

from .. import consts, resources, sh


def setup(cluster_name: str) -> None:
    _create_cluster(cluster_name)
    _create_cluster_role_binding()
    _create_persistent_volume()
    _initialize_helm()
    _helm_add_prometheus_operator()
    _helm_add_prometheus()


def _create_cluster(name: str) -> None:
    logging.info('creating cluster "%s"', name)
    sh.run_gcloud(['container', 'clusters', 'create', name], check=True)
    sh.run_gcloud(
        ['container', 'clusters', 'get-credentials', name], check=True)


def _create_cluster_role_binding() -> None:
    proc = sh.run_gcloud(['config', 'get-value', 'account'], check=True)
    account = proc.stdout
    sh.run_kubectl(
        [
            'create', 'clusterrolebinding', 'cluster-admin-binding',
            '--clusterrole', 'cluster-admin', '--user', account
        ],
        check=True)


def _create_persistent_volume() -> None:
    sh.run_kubectl(
        ['apply', '-f', resources.PERSISTENT_VOLUME_YAML_PATH], check=True)


def _initialize_helm() -> None:
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
    sh.run_helm(
        [
            'install', 'coreos/prometheus-operator', '--name',
            'prometheus-operator', '--namespace', consts.MONITORING_NAMESPACE
        ],
        check=True)


def _helm_add_prometheus() -> None:
    sh.run_helm(
        [
            'install', 'coreos/prometheus', '--name', 'prometheus',
            '--namespace', consts.MONITORING_NAMESPACE, '--values',
            resources.PROMETHEUS_STORAGE_VALUES_YAML_PATH
        ],
        check=True)

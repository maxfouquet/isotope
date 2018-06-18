import logging
import os
import traceback
from typing import Optional, Type

from .. import consts, sh, wait

_RESOURCES_DIR = os.path.realpath(
    os.path.join(os.getcwd(), os.path.dirname(__file__)))

CLIENT_YAML_PATH = os.path.join(_RESOURCES_DIR, 'client.yaml')
HELM_SERVICE_ACCOUNT_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                              'helm-service-account.yaml')
ISTIO_YAML_PATH = os.path.join(_RESOURCES_DIR, 'istio.yaml')
PERSISTENT_VOLUME_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'persistent-volume.yaml')
PROMETHEUS_VALUES_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'values-prometheus.yaml')
SERVICE_GRAPH_YAML_PATH = os.path.join(_RESOURCES_DIR, 'service-graph.yaml')


class Yaml:
    def __init__(self, path: str) -> None:
        self.path = path

    def __enter__(self) -> None:
        _create_from_manifest(self.path)

    def __exit__(self, exception_type: Optional[Type[BaseException]],
                 exception_value: Optional[Exception],
                 traceback: traceback.TracebackException) -> None:
        _delete_from_manifest(self.path)


class NamespacedYaml(Yaml):
    def __init__(self, path: str,
                 namespace: str = consts.DEFAULT_NAMESPACE) -> None:
        super().__init__(path)
        self.namespace = namespace

    def __enter__(self) -> None:
        _create_namespace(self.namespace)
        super().__enter__()

    def __exit__(self, exception_type: Optional[Type[BaseException]],
                 exception_value: Optional[Exception],
                 traceback: traceback.TracebackException) -> None:
        if exception_type is not None:
            logging.error('%s', exception_value)
            logging.info('caught error, exiting')
        super().__exit__(exception_type, exception_value, traceback)
        _delete_namespace(self.namespace)


def _create_namespace(namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    logging.info('creating namespace %s', namespace)
    sh.run_kubectl(['create', 'namespace', namespace], check=True)


def _delete_namespace(namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    logging.info('deleting namespace %s', namespace)
    sh.run_kubectl(['delete', 'namespace', namespace], check=True)
    wait.until(lambda: _namespace_is_deleted(namespace))


def _namespace_is_deleted(namespace: str = consts.DEFAULT_NAMESPACE) -> bool:
    proc = sh.run_kubectl(['get', 'namespace', namespace])
    return proc.returncode != 0


def _create_from_manifest(path: str) -> None:
    logging.info('creating from %s', path)
    sh.run_kubectl(['create', '-f', path], check=True)


def _delete_from_manifest(path: str) -> None:
    logging.info('deleting from %s', path)
    sh.run_kubectl(['delete', '-f', path], check=True)

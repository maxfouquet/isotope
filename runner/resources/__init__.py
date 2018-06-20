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
PROMETHEUS_VALUES_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'values-prometheus.yaml')
PERSISTENT_VOLUME_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'persistent-volume.yaml')
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


def _create_from_manifest(path: str) -> None:
    logging.info('creating from %s', path)
    sh.run_kubectl(['create', '-f', path], check=True)


def _delete_from_manifest(path: str) -> None:
    logging.info('deleting from %s', path)
    sh.run_kubectl(['delete', '-f', path], check=True)

import contextlib
import logging
import os
from typing import Generator, Optional, Type

from .. import consts, sh, wait

_RESOURCES_DIR = os.path.realpath(
    os.path.join(os.getcwd(), os.path.dirname(__file__)))

HELM_SERVICE_ACCOUNT_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                              'helm-service-account.yaml')

STACKDRIVER_PROMETHEUS_GEN_YAML_PATH = os.path.join(
    _RESOURCES_DIR, 'stackdriver-prometheus.gen.yaml')
SERVICE_GRAPH_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'service-graph.gen.yaml')
ISTIO_INGRESS_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                       'istio-ingress.gen.yaml')


@contextlib.contextmanager
def manifest(
        path: str,
        should_tear_down: bool = True,
        should_tear_down_on_error: bool = True) -> Generator[None, None, None]:
    """Runs `kubectl create -f path` on entry and opposing delete on exit."""
    try:
        _create_from_manifest(path)
        yield
    except Exception as e:
        logging.error('%s', e)
        if should_tear_down_on_error:
            _delete_from_manifest(path)
        raise e
    if should_tear_down:
        _delete_from_manifest(path)


def _create_from_manifest(path: str) -> None:
    logging.info('creating from %s', path)
    sh.run_kubectl(['create', '-f', path], check=True)


def _delete_from_manifest(path: str) -> None:
    logging.info('deleting from %s', path)
    sh.run_kubectl(['delete', '-f', path])

import contextlib
import logging
import os
import tempfile
import time
from typing import Any, Dict, Generator

import yaml

from . import consts, kubectl, resources, sh, wait


def set_up(entrypoint_service_name: str, entrypoint_service_namespace: str,
           hub: str, tag: str, should_build: bool) -> None:
    """Installs Istio from master, using hub:tag for the images.

    Requires Helm client to be present.

    This clones the repo in a temporary directory, builds and pushes the
    images, then creates the resources generated via `helm template`.
    """
    with tempfile.TemporaryDirectory() as tmp_go_path:
        repo_path = os.path.join(tmp_go_path, 'src', 'istio.io', 'istio')
        _clone(repo_path)
        if should_build:
            _build_and_push_images(tmp_go_path, repo_path, hub, tag)

        chart_path = os.path.join(repo_path, 'install', 'kubernetes', 'helm',
                                  'istio')
        values_path = os.path.join(chart_path, 'values-isotope.yaml')
        _gen_helm_values(values_path, hub, tag)

        logging.info('installing Helm chart for Istio')
        sh.run_kubectl(['create', 'namespace', consts.ISTIO_NAMESPACE])
        _install(chart_path, values_path, consts.ISTIO_NAMESPACE)

        _create_ingress_rules(entrypoint_service_name,
                              entrypoint_service_namespace)


def get_ingress_gateway_url() -> str:
    ip = wait.until_output([
        'kubectl', '--namespace', consts.ISTIO_NAMESPACE, 'get', 'service',
        'istio-ingressgateway', '-o',
        'jsonpath={.status.loadBalancer.ingress[0].ip}'
    ])
    return 'http://{}:{}'.format(ip, consts.ISTIO_INGRESS_GATEWAY_PORT)


def _clone(path: str) -> None:
    """Clones github.com/istio.io/istio to path."""
    logging.info('cloning istio.io/istio to %s', path)
    sh.run(
        ['git', 'clone', 'https://github.com/istio/istio.git', path],
        check=True)


def _build_and_push_images(go_path: str, repo_path: str, hub: str,
                           tag: str) -> None:
    logging.info('pushing images to %s with tag %s', hub, tag)
    with _work_dir(repo_path):
        env = {
            'GOPATH': go_path,
            'HUB': hub,
            'TAG': tag,
            **dict(os.environ),
        }
        sh.run(['make', 'docker.push'], env=env, check=True)


def _gen_helm_values(path: str, hub: str, tag: str) -> str:
    parent_dir = os.path.dirname(path)
    if not os.path.exists(parent_dir):
        os.makedirs(parent_dir)

    with open(path, 'w') as f:
        return yaml.dump(
            {
                'global': {
                    'hub': hub,
                    'tag': tag,
                }
            },
            f,
            default_flow_style=False)


def _install(chart_path: str,
             values_path: str,
             namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    istio_yaml = sh.run(
        [
            'helm', 'template', chart_path, '--values', values_path,
            '--namespace', namespace
        ],
        check=True).stdout
    kubectl.apply_text(
        istio_yaml, intermediate_file_path=resources.ISTIO_GEN_YAML_PATH)


@contextlib.contextmanager
def _work_dir(path: str) -> Generator[None, None, None]:
    prev_path = os.getcwd()
    if not os.path.exists(path):
        os.makedirs(path)
    os.chdir(path)
    yield
    os.chdir(prev_path)


def _create_ingress_rules(entrypoint_service_name: str,
                          entrypoint_service_namespace: str) -> None:
    logging.info('creating istio ingress rules')
    ingress_yaml = _get_ingress_yaml(entrypoint_service_name,
                                     entrypoint_service_namespace)
    kubectl.apply_text(
        ingress_yaml, intermediate_file_path=resources.ISTIO_INGRESS_YAML_PATH)


def _get_ingress_yaml(entrypoint_service_name: str,
                      entrypoint_service_namespace: str) -> str:
    gateway = _get_gateway_dict()
    virtual_service = _get_virtual_service_dict(entrypoint_service_name,
                                                entrypoint_service_namespace)
    return yaml.dump_all([gateway, virtual_service], default_flow_style=False)


def _get_gateway_dict() -> Dict[str, Any]:
    return {
        'apiVersion': 'networking.istio.io/v1alpha3',
        'kind': 'Gateway',
        'metadata': {
            'name': 'entrypoint-gateway',
        },
        'spec': {
            'selector': {
                'istio': 'ingressgateway',
            },
            'servers': [{
                'hosts': ['*'],
                'port': {
                    'name': 'http',
                    'number': consts.ISTIO_INGRESS_GATEWAY_PORT,
                    'protocol': 'HTTP',
                },
            }],
        },
    }


def _get_virtual_service_dict(
        entrypoint_service_name: str,
        entrypoint_service_namespace: str) -> Dict[str, Any]:
    return {
        'apiVersion': 'networking.istio.io/v1alpha3',
        'kind': 'VirtualService',
        'metadata': {
            'name': 'entrypoint',
        },
        'spec': {
            'hosts': ['*'],
            'gateways': ['entrypoint-gateway'],
            'http': [{
                'route': [{
                    'destination': {
                        'host':
                        '{}.{}.svc.cluster.local'.format(
                            entrypoint_service_name,
                            entrypoint_service_namespace),
                        'port': {
                            'number': consts.SERVICE_PORT,
                        },
                    },
                }],
            }],
        },
    }


def tear_down() -> None:
    """Deletes the Istio resources and namespace."""
    sh.run_kubectl(['delete', '-f', resources.ISTIO_GEN_YAML_PATH])
    sh.run_kubectl(['delete', 'namespace', consts.ISTIO_NAMESPACE])
    wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)

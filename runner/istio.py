import contextlib
import logging
import os
import tempfile
from typing import Any, Dict, Generator

import yaml

from . import consts, context, dicts, resources, sh, wait

_HELM_ISTIO_NAME = 'istio'


@contextlib.contextmanager
def latest(entrypoint_service_name: str, hub: str, tag: str,
           should_build: bool) -> Generator[None, None, None]:
    _install_latest(entrypoint_service_name, hub, tag, should_build)
    with context.confirm_clean_up_on_exception():
        yield
    _clean_up()


def _install_latest(entrypoint_service_name: str, hub: str, tag: str,
                    should_build: bool) -> None:
    """Installs Istio from master, using hub:tag for the images.

    Requires Helm to be present.

    This clones the repo in a temporary directory, builds and pushes the
    images, then runs `helm install`.
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
        _install_helm_chart(chart_path, values_path, _HELM_ISTIO_NAME,
                            consts.ISTIO_NAMESPACE)

        _create_ingress_rules(entrypoint_service_name)


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
        env = dicts.combine(
            dict(os.environ), {
                'GOPATH': go_path,
                'HUB': hub,
                'TAG': tag,
            })
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


def _install_helm_chart(chart_path: str,
                        values_path: str,
                        name: str,
                        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    sh.run_helm(
        [
            'install', chart_path, '--values', values_path, '--name', name,
            '--namespace', namespace
        ],
        check=True)


@contextlib.contextmanager
def _work_dir(path: str) -> Generator[None, None, None]:
    prev_path = os.getcwd()
    if not os.path.exists(path):
        os.makedirs(path)
    os.chdir(path)
    yield
    os.chdir(prev_path)


def _create_ingress_rules(entrypoint_service_name: str) -> None:
    logging.info('creating istio ingress rules')
    ingress_yaml = _get_ingress_yaml(entrypoint_service_name)
    path = resources.ISTIO_INGRESS_YAML_PATH
    with open(path, 'w') as f:
        f.write(ingress_yaml)
    sh.run_kubectl(['create', '-f', path])


def _get_ingress_yaml(entrypoint_service_name: str) -> str:
    gateway = _get_gateway_dict()
    virtual_service = _get_virtual_service_dict(entrypoint_service_name)
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


def _get_virtual_service_dict(entrypoint_service_name: str) -> Dict[str, Any]:
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
                'match': [{
                    'uri': {
                        'prefix': '/',
                    },
                # TODO: is /metrics needed?
                # }, {
                #     'uri': {
                #         'prefix': '/metrics',
                #     },
                }],
                'route': [{
                    'destination': {
                        'port': {
                            'number': consts.SERVICE_PORT,
                        },
                        'host': entrypoint_service_name,
                    },
                }],
            }],
        },
    }


def _clean_up() -> None:
    """Deletes the Istio Helm chart and any leftover resources."""
    sh.run_helm(['delete', '--purge', _HELM_ISTIO_NAME])
    # TODO: Why doesn't `helm delete --purge istio` do this?
    sh.run_kubectl(['delete', 'namespace', consts.ISTIO_NAMESPACE])
    wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)

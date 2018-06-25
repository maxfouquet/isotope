import logging
import os
import time
from typing import Tuple

from . import cluster, consts, istio, resources, sh, wait

_REPO_ROOT = os.path.join(os.getcwd(),
                          os.path.dirname(os.path.dirname(__file__)))
_MAIN_GO_PATH = os.path.join(_REPO_ROOT, 'convert', 'main.go')


def run(topology_path: str, hub: str, tag: str) -> None:
    service_graph_path, prometheus_values_path, client_path = (
        _gen_yaml(topology_path))

    logging.info('updating Prometheus configuration')
    sh.run_helm(
        [
            'upgrade', 'prometheus', 'coreos/prometheus', '--values',
            prometheus_values_path
        ],
        check=True)

    with istio.latest(hub, tag):
        topology_name = _get_basename_no_ext(topology_path)
        _test_service_graph(service_graph_path, client_path,
                            '{}.log'.format(topology_name))


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_yaml(topology_path: str) -> Tuple[str, str, str]:
    logging.info('generating Kubernetes manifests from %s', topology_path)
    client_node_selector = 'cloud.google.com/gke-nodepool={}'.format(
        consts.CLIENT_NODE_POOL_NAME)
    sh.run(
        [
            'go', 'run', _MAIN_GO_PATH, 'kubernetes', topology_path,
            resources.SERVICE_GRAPH_GEN_YAML_PATH,
            resources.PROMETHEUS_VALUES_GEN_YAML_PATH,
            resources.CLIENT_GEN_YAML_PATH, client_node_selector
        ],
        check=True)
    return (resources.SERVICE_GRAPH_GEN_YAML_PATH,
            resources.PROMETHEUS_VALUES_GEN_YAML_PATH,
            resources.CLIENT_GEN_YAML_PATH)


def _test_service_graph(service_graph_path: str, client_path: str,
                        output_path: str) -> None:
    with resources.manifest(service_graph_path):
        wait.until_deployments_are_ready(consts.SERVICE_GRAPH_NAMESPACE)
        wait.until_service_graph_is_ready()
        # TODO: Why is this extra buffer necessary?
        logging.debug('sleeping for 30 seconds as an extra buffer')
        time.sleep(30)

        with resources.manifest(client_path):
            wait.until_client_job_is_complete()
            _write_job_logs(output_path, consts.CLIENT_JOB_NAME)
            wait.until_prometheus_has_scraped()

    wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)


def _write_job_logs(path: str, job_name: str) -> None:
    logging.info('retrieving logs for %s', job_name)
    # TODO: get logs for each pod in job
    # TODO: get logs for the successful pod in job
    proc = sh.run_kubectl(['logs', 'job/{}'.format(job_name)], check=True)
    logs = proc.stdout
    _write_to_file(path, logs)


def _write_to_file(path: str, contents: str) -> None:
    logging.debug('writing contents to %s', path)
    with open(path, 'w') as f:
        f.writelines(contents)

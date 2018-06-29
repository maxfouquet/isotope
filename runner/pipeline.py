import contextlib
import logging
import os
import time
from typing import Dict, Generator, Tuple

from . import cluster, config, consts, dicts, istio, md5, prometheus, \
              resources, sh, wait

_REPO_ROOT = os.path.join(os.getcwd(),
                          os.path.dirname(os.path.dirname(__file__)))
_MAIN_GO_PATH = os.path.join(_REPO_ROOT, 'convert', 'main.go')


def run(topology_path: str, environment: config.Environment,
        service_image: str, client_image: str, client_args: str, hub: str,
        tag: str, should_build_istio: bool,
        static_labels: Dict[str, str]) -> None:
    service_graph_path, client_path = _gen_yaml(topology_path, service_image,
                                                client_image, client_args)

    topology_name = _get_basename_no_ext(topology_path)
    _update_prometheus_configuration(topology_path, environment, topology_name,
                                     static_labels)

    if environment == config.Environment.NONE:
        environment_setup = _no_op
    else:
        environment_setup = lambda: istio.latest(hub, tag, should_build_istio)

    with environment_setup():
        env_name = environment.name.lower()
        logging.info('starting test with environment "%s"', env_name)
        _test_service_graph(service_graph_path, client_path,
                            '{}_{}.log'.format(topology_name, env_name))


@contextlib.contextmanager
def _no_op() -> Generator[None, None, None]:
    yield


def _update_prometheus_configuration(
        topology_path: str, environment: config.Environment,
        topology_name: str, static_labels: Dict[str, str]) -> None:
    _write_prometheus_values_for_topology(topology_path, environment,
                                          topology_name, static_labels)

    logging.info('updating Prometheus configuration')
    sh.run_helm(
        [
            'upgrade', 'prometheus', 'coreos/prometheus', '--values',
            resources.PROMETHEUS_VALUES_GEN_YAML_PATH
        ],
        check=True)


def _write_prometheus_values_for_topology(
        path: str, environment: config.Environment, name: str,
        labels: Dict[str, str]) -> None:
    labels = dicts.combine(
        labels, {
            'environment': environment.name,
            'topology_name': name,
            'topology_hash': md5.hex(path),
        })
    _write_prometheus_values(labels)


def _write_prometheus_values(labels: Dict[str, str]) -> None:
    values_yaml = prometheus.values_yaml(labels)
    with open(resources.PROMETHEUS_VALUES_GEN_YAML_PATH, 'w') as f:
        f.write(values_yaml)


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_yaml(topology_path: str, service_image: str, client_image: str,
              client_args: str) -> Tuple[str, str]:
    logging.info('generating Kubernetes manifests from %s', topology_path)
    service_graph_node_selector = _get_gke_node_selector(
        consts.SERVICE_GRAPH_NODE_POOL_NAME)
    client_node_selector = _get_gke_node_selector(consts.CLIENT_NODE_POOL_NAME)
    sh.run(
        [
            'go', 'run', _MAIN_GO_PATH, 'kubernetes', '--service-image',
            service_image, '--client-image', client_image, '--client-args',
            client_args, topology_path, resources.SERVICE_GRAPH_GEN_YAML_PATH,
            resources.CLIENT_GEN_YAML_PATH, service_graph_node_selector,
            client_node_selector
        ],
        check=True)
    return (resources.SERVICE_GRAPH_GEN_YAML_PATH,
            resources.CLIENT_GEN_YAML_PATH)


def _get_gke_node_selector(node_pool_name: str) -> str:
    return 'cloud.google.com/gke-nodepool={}'.format(node_pool_name)


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

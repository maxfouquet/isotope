import contextlib
import logging
import os
import time
from typing import Dict, Generator, Optional, Tuple

import requests

from . import cluster, consts, dicts, entrypoint, istio, md5, mesh, \
              prometheus, resources, sh, wait

_REPO_ROOT = os.path.join(os.getcwd(),
                          os.path.dirname(os.path.dirname(__file__)))
_MAIN_GO_PATH = os.path.join(_REPO_ROOT, 'convert', 'main.go')


def run(topology_path: str, env: mesh.Environment, should_tear_down: bool,
        should_tear_down_on_error: bool, cluster_project_id: str,
        cluster_name: str, cluster_zone: str, service_image: str,
        client_image: str, hub: str, tag: str, should_build_istio: bool,
        test_qps: Optional[int], test_duration: str,
        test_num_concurrent_connections: int,
        static_labels: Dict[str, str]) -> None:
    """Runs a load test on the topology in topology_path with the environment.

    Args:
        topology_path: the path to the file containing the topology
        env: the pre-existing mesh environment for the topology (i.e. Istio)
        should_tear_down: if True, calls env.tear_down() after testing
        should_tear_down_on_error: if True, calls env.tear_down() on errors
        service_image: the Docker image to represent each node in the topology
        client_image: the Docker image which can run a load test (i.e. Fortio)
        hub: the Docker hub for Istio images
        tag: the image tag for Istio images
        should_build_istio: if True, builds and pushes Istio images from master
        test_qps: the target QPS for the client; None = max
        test_duration: the duration for the client to run
        test_num_concurrent_connections: the number of simultaneous connections
                for the client to make
        static_labels: labels to add to each Prometheus metric
    """

    manifest_path = _gen_yaml(topology_path, service_image, client_image)

    topology_name = _get_basename_no_ext(topology_path)
    labels = dicts.combine(
        static_labels, {
            'environment': env.name,
            'topology_name': topology_name,
            'topology_hash': md5.hex(topology_path),
        })
    prometheus.apply(cluster_project_id, cluster_name, cluster_zone, labels)

    with env.context(
            should_tear_down=should_tear_down,
            should_tear_down_on_error=should_tear_down_on_error
    ) as ingress_url:
        logging.info('starting test with environment "%s"', env.name)
        result_output_path = '{}_{}.json'.format(topology_name, env.name)

        _test_service_graph(manifest_path, should_tear_down,
                            should_tear_down_on_error, result_output_path,
                            ingress_url, test_qps, test_duration,
                            test_num_concurrent_connections)


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_yaml(topology_path: str, service_image: str,
              client_image: str) -> str:
    """Converts topology_path to Kubernetes manifests.

    The neighboring Go command in convert/ handles this operation.

    Args:
        topology_path: the path containing the topology YAML
        service_image: the Docker image to represent each node in the topology;
                passed to the Go command
        client_image: the Docker image which can run a load test (i.e. Fortio);
                passed to the Go command
    """
    logging.info('generating Kubernetes manifests from %s', topology_path)
    service_graph_node_selector = _get_gke_node_selector(
        consts.SERVICE_GRAPH_NODE_POOL_NAME)
    client_node_selector = _get_gke_node_selector(consts.CLIENT_NODE_POOL_NAME)
    sh.run(
        [
            'go', 'run', _MAIN_GO_PATH, 'kubernetes', '--service-image',
            service_image, '--client-image', client_image, topology_path,
            resources.SERVICE_GRAPH_GEN_YAML_PATH, service_graph_node_selector,
            client_node_selector
        ],
        check=True)
    return resources.SERVICE_GRAPH_GEN_YAML_PATH


def _get_gke_node_selector(node_pool_name: str) -> str:
    return 'cloud.google.com/gke-nodepool={}'.format(node_pool_name)


def _test_service_graph(yaml_path: str, should_tear_down: bool,
                        should_tear_down_on_error: bool,
                        test_result_output_path: str, test_target_url: str,
                        test_qps: Optional[int], test_duration: str,
                        test_num_concurrent_connections: int) -> None:
    """Deploys the service graph at yaml_path and runs a load test on it."""
    with resources.manifest(
            yaml_path,
            should_tear_down=should_tear_down,
            should_tear_down_on_error=should_tear_down_on_error):
        wait.until_deployments_are_ready(consts.SERVICE_GRAPH_NAMESPACE)
        wait.until_service_graph_is_ready()
        # TODO: Why is this extra buffer necessary?
        logging.debug('sleeping for 30 seconds as an extra buffer')
        time.sleep(30)

        _run_load_test(test_result_output_path, test_target_url, test_qps,
                       test_duration, test_num_concurrent_connections)

        wait.until_prometheus_has_scraped()

    wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)


def _run_load_test(result_output_path: str, test_target_url: str,
                   test_qps: Optional[int], test_duration: str,
                   test_num_concurrent_connections: int) -> None:
    """Sends an HTTP request to the client; expecting a JSON response.

    The HTTP request's query string contains the necessary info to perform
    the load test, adapted from the arguments described in
    https://github.com/istio/istio/blob/master/tools/README.md#run-the-functions.

    Args:
        result_output_path: the path to write the JSON output.
        test_target_url: the in-cluster URL to
        test_qps: the target QPS for the client; None = max
        test_duration: the duration for the client to run
        test_num_concurrent_connections: the number of simultaneous connections
                for the client to make
    """
    logging.info('starting load test')
    svc_addr = _get_svc_ip(consts.CLIENT_NAME)
    qps = -1 if test_qps is None else test_qps  # -1 indicates max QPS.
    url = ('http://{}:{}/fortio?json=on&qps={}&t={}&c={}&load=Start&url={}'
           ).format(svc_addr, consts.CLIENT_PORT, qps, test_duration,
                    test_num_concurrent_connections, test_target_url)
    result = _http_get_json(url)
    _write_to_file(result_output_path, result)


def _get_svc_ip(name: str) -> str:
    """Blocks until a public IP address for name is created, and returns it."""
    ip = None
    while ip is None:
        output = sh.run_kubectl([
            'get', 'service', name, '-o',
            'jsonpath={.status.loadBalancer.ingress[0].ip}'
        ]).stdout
        if output:
            ip = output
        else:
            logging.debug(
                'waiting for service/%s to obtain an external IP address',
                name)
            time.sleep(wait.RETRY_INTERVAL.seconds)
    logging.debug('service/%s IP is %s', name, ip)
    return ip


def _http_get_json(url: str) -> str:
    """Sends an HTTP GET request to url, returning its JSON response."""
    response = None
    while response is None:
        try:
            response = requests.get(url)
        except (requests.ConnectionError, requests.HTTPError) as e:
            logging.error('%s; retrying request to %s', e, url)
            response = None
    return response.text


def _write_to_file(path: str, contents: str) -> None:
    logging.debug('writing contents to %s', path)
    with open(path, 'w') as f:
        f.writelines(contents)

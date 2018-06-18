import logging
import os
import time
from typing import Tuple

from runner import cluster, consts, resources, sh, wait

_MAIN_GO_PATH = os.path.realpath(
    os.path.join(os.getcwd(),
                 os.path.dirname(os.path.dirname(os.path.dirname(__file__))),
                 'main.go'))


def run(topology_path: str) -> None:
    service_graph_path, client_path = _gen_yaml(topology_path)

    base_name_no_ext = _get_basename_no_ext(topology_path)

    _test_service_graph(service_graph_path, client_path,
                        '{}_no-istio.log'.format(base_name_no_ext))

    _test_service_graph_with_istio(
        resources.ISTIO_YAML_PATH, service_graph_path, client_path,
        '{}_with-istio.log'.format(base_name_no_ext))


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_yaml(topology_path: str) -> Tuple[str, str]:
    logging.info('generating Kubernetes manifests from %s', topology_path)
    sh.run_cmd(
        [
            'go', 'run', _MAIN_GO_PATH, 'performance', 'kubernetes',
            topology_path, resources.SERVICE_GRAPH_YAML_PATH,
            resources.CLIENT_YAML_PATH
        ],
        check=True)
    return resources.SERVICE_GRAPH_YAML_PATH, resources.CLIENT_YAML_PATH


def _test_service_graph(service_graph_path: str, client_path: str,
                        output_path: str) -> None:
    with resources.NamespacedYaml(service_graph_path,
                                  consts.SERVICE_GRAPH_NAMESPACE):
        wait.until_deployments_are_ready(consts.SERVICE_GRAPH_NAMESPACE)
        wait.until_service_graph_is_ready()
        # TODO: Why is this extra buffer necessary?
        logging.debug('sleeping for 30 seconds as an extra buffer')
        time.sleep(30)

        with resources.Yaml(client_path):
            wait.until_client_job_is_complete()
            _write_job_logs(output_path, consts.CLIENT_JOB_NAME)
            wait.until_prometheus_has_scraped()


def _test_service_graph_with_istio(istio_path: str, service_graph_path: str,
                                   client_path: str, output_path: str) -> None:
    with resources.NamespacedYaml(istio_path, consts.ISTIO_NAMESPACE):
        wait.until_deployments_are_ready(consts.ISTIO_NAMESPACE)

        _test_service_graph(service_graph_path, client_path, output_path)


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

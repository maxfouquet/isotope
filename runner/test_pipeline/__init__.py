import logging
import os
import shutil
import tempfile
import time
import traceback
from typing import Optional, Tuple, Type

from runner import cluster, consts, resources, sh, wait

_MAIN_GO_PATH = os.path.realpath(
    os.path.join(os.getcwd(),
                 os.path.dirname(os.path.dirname(os.path.dirname(__file__))),
                 'main.go'))


def run(topology_path: str, istio_path: str = None) -> None:
    service_graph_path, prometheus_values_path, client_path = (
        _gen_yaml(topology_path))

    logging.info('updating Prometheus configuration')
    sh.run_helm(
        [
            'upgrade', 'prometheus', 'coreos/prometheus', '--values',
            prometheus_values_path
        ],
        check=True)

    def test() -> None:
        topology_name = _get_basename_no_ext(topology_path)
        istio_name = _get_basename_no_ext(istio_path) if istio_path else None
        _test_service_graph(service_graph_path, client_path,
                            '{}_{}.log'.format(topology_name, istio_name))

    if istio_path is None:
        test()
    else:
        with Istio():
            test()


class Istio:
    def __init__(self) -> None:
        # self._tmp_dir = tempfile.mkdtemp()
        self._tmp_dir = '.'
        self.path = os.path.join(self._tmp_dir, 'istio')

    def __enter__(self) -> None:
        _clone_istio_repo(self.path)
        _install_istio_helm_chart(self.path)

    def __exit__(self, exception_type: Optional[Type[BaseException]],
                 exception_value: Optional[Exception],
                 traceback: traceback.TracebackException) -> None:
        _delete_istio_helm_chart()
        # shutil.rmtree(self._tmp_dir)


def _clone_istio_repo(path: str) -> None:
    """Clones github.com/istio.io/istio to the path."""
    logging.info('cloning istio.io/istio to %s', path)
    sh.run_cmd(
        ['git', 'clone', 'https://github.com/istio/istio.git', path],
        check=True)


def _install_istio_helm_chart(path: str) -> None:
    logging.info('installing Helm chart for Istio')
    helm_chart_path = os.path.join(path, 'install', 'kubernetes', 'helm',
                                   'istio')
    sh.run_helm(
        [
            'install', helm_chart_path, '--name', 'istio', '--namespace',
            consts.ISTIO_NAMESPACE
        ],
        check=True)


def _delete_istio_helm_chart() -> None:
    sh.run_helm(['delete', '--purge', 'istio'])


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_yaml(topology_path: str) -> Tuple[str, str, str]:
    logging.info('generating Kubernetes manifests from %s', topology_path)
    sh.run_cmd(
        [
            'go', 'run', _MAIN_GO_PATH, 'performance', 'kubernetes',
            topology_path, resources.SERVICE_GRAPH_YAML_PATH,
            resources.PROMETHEUS_VALUES_YAML_PATH, resources.CLIENT_YAML_PATH
        ],
        check=True)
    return (resources.SERVICE_GRAPH_YAML_PATH,
            resources.PROMETHEUS_VALUES_YAML_PATH, resources.CLIENT_YAML_PATH)


def _test_service_graph(service_graph_path: str, client_path: str,
                        output_path: str) -> None:
    with resources.Yaml(service_graph_path):
        wait.until_deployments_are_ready(consts.SERVICE_GRAPH_NAMESPACE)
        wait.until_service_graph_is_ready()
        # TODO: Why is this extra buffer necessary?
        logging.debug('sleeping for 30 seconds as an extra buffer')
        time.sleep(30)

        with resources.Yaml(client_path):
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

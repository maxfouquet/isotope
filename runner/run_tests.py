#!/usr/bin/env python3

import argparse
import datetime
import os
import json
import logging
import subprocess
import sys
import time
import traceback
from typing import Any, Callable, List, Optional, Tuple, Type

DEFAULT_NAMESPACE = 'default'
SERVICE_GRAPH_NAMESPACE = 'service-graph'
SERVICE_GRAPH_SERVICE_SELECTOR = 'role=service'
CLIENT_JOB_NAME = 'client'
ISTIO_NAMESPACE = 'istio-system'

RETRY_INTERVAL = datetime.timedelta(seconds=5)


def main() -> None:
    args = parse_args()
    log_level = getattr(logging, args.log_level)
    logging.basicConfig(level=log_level, format='%(levelname)s\t> %(message)s')

    for topology_path in args.topology_paths:
        service_graph_path, client_path = gen_yaml(topology_path)

        base_name_no_ext = get_basename_no_ext(topology_path)

        test_service_graph(service_graph_path, client_path,
                           '{}_no-istio.log'.format(base_name_no_ext))

        test_service_graph_with_istio(
            'istio.yaml', service_graph_path, client_path,
            '{}_with-istio.log'.format(base_name_no_ext))


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument('topology_paths', metavar='PATH', type=str, nargs='+')
    parser.add_argument(
        '--log_level',
        type=str,
        choices=['CRITICAL', 'ERROR', 'WARNING', 'INFO', 'DEBUG'],
        default='INFO')
    return parser.parse_args()


def get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def gen_yaml(topology_path: str) -> Tuple[str, str]:
    logging.info('generating Kubernetes manifests from %s', topology_path)
    run_cmd([
        'go', 'run', '../main.go', 'performance', 'kubernetes', topology_path
    ])
    return 'service-graph.yaml', 'client.yaml'


def test_service_graph(service_graph_path: str, client_path: str,
                       output_path: str) -> None:
    with NamespacedYamlResources(service_graph_path, SERVICE_GRAPH_NAMESPACE):
        block_until_deployments_are_ready(SERVICE_GRAPH_NAMESPACE)
        block_until(service_graph_is_ready)
        # TODO: Why is this extra buffer necessary?
        logging.debug('sleeping for 30 seconds as an extra buffer')
        time.sleep(30)

        with YamlResources(client_path):
            block_until(client_job_is_complete)
            write_job_logs(output_path, CLIENT_JOB_NAME)


class YamlResources:
    def __init__(self, path: str) -> None:
        self.path = path

    def __enter__(self) -> None:
        create_from_manifest(self.path)

    def __exit__(self, exception_type: Optional[Type[BaseException]],
                 exception_value: Optional[Exception],
                 traceback: traceback.TracebackException) -> None:
        delete_from_manifest(self.path)


class NamespacedYamlResources(YamlResources):
    def __init__(self, path: str, namespace: str = DEFAULT_NAMESPACE) -> None:
        super().__init__(path)
        self.namespace = namespace

    def __enter__(self) -> None:
        create_namespace(self.namespace)
        super().__enter__()

    def __exit__(self, exception_type: Optional[Type[BaseException]],
                 exception_value: Optional[Exception],
                 traceback: traceback.TracebackException) -> None:
        super().__exit__(exception_type, exception_value, traceback)
        delete_namespace(self.namespace)


def test_service_graph_with_istio(istio_path: str, service_graph_path: str,
                                  client_path: str, output_path: str) -> None:
    with NamespacedYamlResources(istio_path, ISTIO_NAMESPACE):
        block_until_deployments_are_ready(ISTIO_NAMESPACE)

        test_service_graph(service_graph_path, client_path, output_path)


def write_job_logs(path: str, job_name: str) -> None:
    logging.info('retrieving logs for %s', job_name)
    # TODO: get logs for each pod in job
    # TODO: get logs for the successful pod in job
    logs = check_run_kubectl(['logs', 'job/{}'.format(job_name)])
    write_to_file(path, logs)


def write_to_file(path: str, contents: str) -> None:
    logging.debug('writing contents to %s', path)
    with open(path, 'w') as f:
        f.writelines(contents)


def create_namespace(namespace: str = DEFAULT_NAMESPACE) -> None:
    logging.info('creating namespace %s', namespace)
    check_run_kubectl(['create', 'namespace', namespace])


def delete_namespace(namespace: str = DEFAULT_NAMESPACE) -> None:
    logging.info('deleting namespace %s', namespace)
    check_run_kubectl(['delete', 'namespace', namespace])
    block_until(lambda: namespace_is_deleted(namespace))


def namespace_is_deleted(namespace: str = DEFAULT_NAMESPACE) -> bool:
    _, _, status = run_kubectl(['get', 'namespace', namespace])
    return status != 0


def create_from_manifest(path: str) -> None:
    logging.info('creating from %s', path)
    check_run_kubectl(['create', '-f', path])


def service_graph_is_ready() -> bool:
    out = check_run_kubectl([
        '--namespace', SERVICE_GRAPH_NAMESPACE, 'get', 'pods', '--selector',
        SERVICE_GRAPH_SERVICE_SELECTOR, '-o',
        'jsonpath={.items[*].status.conditions[?(@.type=="Ready")].status}'
    ])
    all_services_ready = out != '' and 'False' not in out
    return all_services_ready


def client_job_is_complete() -> bool:
    out = check_run_kubectl([
        'get', 'job', CLIENT_JOB_NAME, '-o',
        'jsonpath={.status.conditions[?(@.type=="Complete")].status}'
    ])
    return 'True' in out


def block_until_deployments_are_ready(
        namespace: str = DEFAULT_NAMESPACE) -> None:
    out = check_run_kubectl([
        '--namespace', namespace, 'get', 'deployments', '-o',
        'jsonpath={.items[*].metadata.name}'
    ])
    deployments = out.split(' ')
    logging.info('waiting for deployments in %s (%s) to rollout', namespace,
                 deployments)
    for deployment in deployments:
        # kubectl blocks until ready.
        check_run_kubectl([
            '--namespace', namespace, 'rollout', 'status', 'deployment',
            deployment
        ])


def block_until(predicate: Callable[[], bool]) -> None:
    while not predicate():
        time.sleep(RETRY_INTERVAL.seconds)


def delete_from_manifest(path: str) -> None:
    logging.info('deleting from %s', path)
    check_run_kubectl(['delete', '-f', path])


def run_kubectl(args: List[str]) -> Tuple[str, str, int]:
    return run_cmd(['kubectl', *args])


def check_run_kubectl(args: List[str]) -> str:
    try:
        out, _, _ = run_cmd(['kubectl', *args], check=True)
        return out
    except subprocess.CalledProcessError as e:
        logging.error('process exited with code %d\n'
                      '%s\n', e.returncode, e.stderr)
        raise e


def run_cmd(params: List[str], input: str = None,
            **kwargs: Any) -> Tuple[str, str, int]:
    logging.debug('%s', ' '.join(params))

    proc = subprocess.run(
        params,
        input=input,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        **kwargs)

    out = proc.stdout.decode('utf-8')
    if out:
        logging.debug('STDOUT: %s', out)

    err = proc.stderr.decode('utf-8')
    if err:
        logging.debug('STDERR: %s', err)

    return out, err, proc.returncode


if __name__ == '__main__':
    main()

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

from runner import cluster, consts, resources, sh, wait

CLUSTER_NAME = 'isotope-cluster'


def main() -> None:
    args = parse_args()
    log_level = getattr(logging, args.log_level)
    logging.basicConfig(level=log_level, format='%(levelname)s\t> %(message)s')

    if args.create_cluster:
        cluster.setup(CLUSTER_NAME)

    for topology_path in args.topology_paths:
        run_test(topology_path)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument('topology_paths', metavar='PATH', type=str, nargs='+')
    parser.add_argument('--create_cluster', default=False, action='store_true')
    parser.add_argument(
        '--log_level',
        type=str,
        choices=['CRITICAL', 'ERROR', 'WARNING', 'INFO', 'DEBUG'],
        default='INFO')
    return parser.parse_args()


def run_test(topology_path: str) -> None:
    service_graph_path, client_path = gen_yaml(topology_path)

    base_name_no_ext = get_basename_no_ext(topology_path)

    test_service_graph(service_graph_path, client_path,
                       '{}_no-istio.log'.format(base_name_no_ext))

    test_service_graph_with_istio(resources.ISTIO_YAML_PATH,
                                  service_graph_path, client_path,
                                  '{}_with-istio.log'.format(base_name_no_ext))


def get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def gen_yaml(topology_path: str) -> Tuple[str, str]:
    logging.info('generating Kubernetes manifests from %s', topology_path)
    sh.run_cmd(
        # TODO: main.go is relative to the repo root, not the WD.
        [
            'go', 'run', 'main.go', 'performance', 'kubernetes', topology_path,
            resources.SERVICE_GRAPH_YAML_PATH, resources.CLIENT_YAML_PATH
        ],
        check=True)
    return resources.SERVICE_GRAPH_YAML_PATH, resources.CLIENT_YAML_PATH


def test_service_graph(service_graph_path: str, client_path: str,
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
            write_job_logs(output_path, consts.CLIENT_JOB_NAME)
            wait.until_prometheus_has_scraped()


def test_service_graph_with_istio(istio_path: str, service_graph_path: str,
                                  client_path: str, output_path: str) -> None:
    with resources.NamespacedYaml(istio_path, consts.ISTIO_NAMESPACE):
        wait.until_deployments_are_ready(consts.ISTIO_NAMESPACE)

        test_service_graph(service_graph_path, client_path, output_path)


def write_job_logs(path: str, job_name: str) -> None:
    logging.info('retrieving logs for %s', job_name)
    # TODO: get logs for each pod in job
    # TODO: get logs for the successful pod in job
    proc = sh.run_kubectl(['logs', 'job/{}'.format(job_name)], check=True)
    logs = proc.stdout
    write_to_file(path, logs)


def write_to_file(path: str, contents: str) -> None:
    logging.debug('writing contents to %s', path)
    with open(path, 'w') as f:
        f.writelines(contents)


if __name__ == '__main__':
    main()

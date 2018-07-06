import collections
import datetime
import logging
import subprocess
import time
from typing import Callable

from . import consts, sh

PROMETHEUS_SCRAPE_INTERVAL = datetime.timedelta(seconds=30)
RETRY_INTERVAL = datetime.timedelta(seconds=5)


def until(predicate: Callable[[], bool]) -> None:
    while not predicate():
        time.sleep(RETRY_INTERVAL.seconds)


def until_deployments_are_ready(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    proc = sh.run_kubectl(
        [
            '--namespace', namespace, 'get', 'deployments', '-o',
            'jsonpath={.items[*].metadata.name}'
        ],
        check=True)
    deployments = collections.deque(proc.stdout.split(' '))
    logging.info('waiting for deployments in %s (%s) to rollout', namespace,
                 deployments)
    while len(deployments) > 0:
        deployment = deployments.popleft()
        # kubectl blocks until ready.
        try:
            sh.run_kubectl(
                [
                    '--namespace', namespace, 'rollout', 'status',
                    'deployment', deployment
                ],
                check=True)
        except subprocess.CalledProcessError as e:
            msg = 'failed to check rollout status of {}'.format(deployment)
            if 'watch closed' in e.stderr:
                logging.debug('%s; retrying later', msg)
                deployments.append(deployment)
            else:
                logging.error(msg)


def until_prometheus_has_scraped() -> None:
    logging.info('allowing Prometheus time to scrape final metrics')
    time.sleep(PROMETHEUS_SCRAPE_INTERVAL.seconds)


def until_namespace_is_deleted(
        namespace: str = consts.DEFAULT_NAMESPACE) -> None:
    until(lambda: _namespace_is_deleted(namespace))


def _namespace_is_deleted(namespace: str = consts.DEFAULT_NAMESPACE) -> bool:
    proc = sh.run_kubectl(['get', 'namespace', namespace])
    return proc.returncode != 0


def until_service_graph_is_ready() -> None:
    until(_service_graph_is_ready)


def _service_graph_is_ready() -> bool:
    proc = sh.run_kubectl(
        [
            '--namespace', consts.SERVICE_GRAPH_NAMESPACE, 'get', 'pods',
            '--selector', consts.SERVICE_GRAPH_SERVICE_SELECTOR, '-o',
            'jsonpath={.items[*].status.conditions[?(@.type=="Ready")].status}'
        ],
        check=True)
    out = proc.stdout
    all_services_ready = out != '' and 'False' not in out
    return all_services_ready

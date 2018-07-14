import contextlib
import logging
import socket
import subprocess
import time
from typing import Generator

from . import sh


@contextlib.contextmanager
def manifest(path: str) -> Generator[None, None, None]:
    """Runs `kubectl apply -f path` on entry and opposing delete on exit."""
    try:
        apply_file(path)
        yield
    finally:
        delete_file(path)


def apply_file(path: str) -> None:
    logging.info('applying from %s', path)
    sh.run_kubectl(['apply', '-f', path], check=True)


def delete_file(path: str) -> None:
    logging.info('deleting from %s', path)
    sh.run_kubectl(['delete', '-f', path])


@contextlib.contextmanager
def port_forward(deployment_name: str, deployment_port: int,
                 namespace: str) -> Generator[int, None, None]:
    """Port forwards to a deployment, yielding the chosen open port."""
    local_port = _get_open_port()
    proc = subprocess.Popen(
        [
            'kubectl', 'port-forward',
            'deployment/{}'.format(deployment_name), '{}:{}'.format(
                local_port, deployment_port), '--namespace', namespace
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE)

    # TODO: Should wait for output from proc.stdout
    time.sleep(1)

    yield local_port

    proc.terminate()


# Adapted from
# https://stackoverflow.com/questions/2838244/get-open-tcp-port-in-python.
def _get_open_port() -> int:
    sock = socket.socket()
    sock.bind(('', 0))
    _, port = sock.getsockname()
    return port

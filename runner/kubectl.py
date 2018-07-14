import contextlib
import socket
import subprocess
import time
from typing import Generator

from . import sh


def apply_file(path: str) -> None:
    sh.run_kubectl(['apply', '-f', path], check=True)


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

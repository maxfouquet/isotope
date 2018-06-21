import logging
import subprocess
from typing import List


def run_gcloud(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['gcloud', *args], check=check)


def run_kubectl(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['kubectl', *args], check=check)


def run_helm(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['helm', *args], check=check)


def run(args: List[str], check=False) -> subprocess.CompletedProcess:
    logging.debug('%s', args)
    try:
        proc = subprocess.run(
            args, check=check, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    except subprocess.CalledProcessError as e:
        logging.error('%s\n%s', e, e.stderr)
        raise e

    if proc.stdout is not None:
        proc.stdout = proc.stdout.decode('utf-8').strip()
    if proc.stderr is not None:
        proc.stderr = proc.stderr.decode('utf-8').strip()
    return proc

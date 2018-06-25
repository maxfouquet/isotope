import logging
import subprocess
from typing import Dict, List, Union


def run_gcloud(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['gcloud', *args], check=check)


def run_kubectl(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['kubectl', *args], check=check)


def run_helm(args: List[str], check=False) -> subprocess.CompletedProcess:
    return run(['helm', *args], check=check)


def run(args: List[str], check=False,
        env: Dict[str, str] = None) -> subprocess.CompletedProcess:
    logging.debug('%s', args)

    try:
        proc = subprocess.run(
            args,
            check=check,
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE)
    except subprocess.CalledProcessError as e:
        _decode(e)
        logging.error('%s\n%s\n%s', e, e.stdout, e.stderr)
        raise e

    _decode(proc)
    return proc


def _decode(
        proc: Union[subprocess.CompletedProcess, subprocess.CalledProcessError]
) -> None:
    if proc.stdout is not None:
        proc.stdout = proc.stdout.decode('utf-8').strip()
    if proc.stderr is not None:
        proc.stderr = proc.stderr.decode('utf-8').strip()

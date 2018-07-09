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
    """Delegates to subprocess.run, capturing stdout and stderr.

    Args:
        args: the list of args, with the command as the first item
        check: if True, raises an exception if the command returns non-zero
        env: the environment variables to set during the command's runtime

    Returns:
        A completed process, with stdout and stderr decoded as UTF-8 strings.
    """
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

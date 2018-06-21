import logging
import subprocess
import sys
from typing import Dict, List


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
            args, check=check, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    except subprocess.CalledProcessError as e:
        logging.error('%s\n%s', e, e.stderr)
        raise e

    if proc.stdout is not None:
        proc.stdout = proc.stdout.decode('utf-8').strip()
    if proc.stderr is not None:
        proc.stderr = proc.stderr.decode('utf-8').strip()
    return proc


# def run(args: List[str], check=False,
#         env: Dict[str, str] = None) -> subprocess.CompletedProcess:
#     logging.debug('%s', args)
#     try:
#         proc = subprocess.Popen(
#             # proc = subprocess.run(
#             args,
#             check=check,
#             env=env,
#             stdout=subprocess.PIPE)
#         if logging.getLogger().isEnabledFor(logging.DEBUG):
#             for line in iter(proc.stdout.readline, b''):
#                 sys.stdout.write(line)

#     except subprocess.CalledProcessError as e:
#         logging.error('%s\n%s', e, e.stderr)
#         raise e

#     if proc.stdout is not None:
#         proc.stdout = proc.stdout.decode('utf-8').strip()
#     if proc.stderr is not None:
#         proc.stderr = proc.stderr.decode('utf-8').strip()
#     return proc

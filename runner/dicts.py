from typing import Dict


def combine(*args: Dict[str, str]) -> Dict[str, str]:
    """Adds all keys from all args into a combined dictionary."""
    acc = {}
    for d in args:
        acc.update(d)
    return acc

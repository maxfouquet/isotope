import contextlib
import functools
import logging
from typing import Generator

_NO_CHAR = 'n'
_YES_CHAR = 'y'
_PROMPT_CHAR = '>'


def _confirm(msg: str, default: bool = None) -> bool:
    yes_char = _YES_CHAR
    no_char = _NO_CHAR
    if default is True:
        yes_char = yes_char.upper()
    elif default is False:
        no_char = no_char.upper()

    choices = '{}/{}'.format(yes_char, no_char)
    prompt = '{} ({}):\n{} '.format(msg, choices, _PROMPT_CHAR)

    choice = None
    while choice is None:
        line = input(prompt).lower()
        if line == '' and default is not None:
            choice = default
        elif line == 'y':
            choice = True
        elif line == 'n':
            choice = False
    return choice


@contextlib.contextmanager
def confirm_clean_up_on_exception() -> Generator[None, None, None]:
    try:
        yield
    except (KeyboardInterrupt, Exception) as e:
        logging.error('%s', e)
        should_continue = _confirm(
            'Caught error {}. Do you want to clean up?'.format(e),
            default=True)
        if not should_continue:
            raise e

import re

_CAMEL_CASE_PATTERN = re.compile(r'(.)([A-Z][a-z]+)')
_SNAKE_CASE_PATTERN = re.compile('([a-z0-9])([A-Z])')


def snakify(s: str) -> str:
    subbed = _CAMEL_CASE_PATTERN.sub(r'\1_\2', s)
    return _SNAKE_CASE_PATTERN.sub(r'\1_\2', subbed).lower()


def snakify_dict(d: dict) -> dict:
    snake_case_dict = {}
    for k, v in d.items():
        k = snakify(k)
        if type(v) is dict:
            v = snakify_dict(v)
        # elif type(v) is list:
        #     for sub_v in v:
        #         if type(v)
        snake_case_dict[k] = v
    return snake_case_dict

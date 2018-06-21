import enum
import logging
from typing import Any, Dict, List

import yaml

from . import case


class Command:
    def __init__(self, name: str) -> None:
        self.name = name


class SleepCommand(Command):
    def __init__(self, seconds: int) -> None:
        self.seconds = seconds


class CallCommand(Command):
    def __init__(self, service: str, size: int = 0) -> None:
        self.service = service
        self.size = size


def _command_from_dict(d: Dict[str, Any]) -> Command:
    if len(d) > 1:
        raise ValueError('a command must not have multiple keys')

    key, value = next(iter(d.items()))
    value_type = type(value)
    if key == 'sleep':
        if value_type is int:
            command = SleepCommand(seconds=value)  # type: Command
        else:
            raise TypeError('SleepCommand requires an int value')
    elif key == 'call':
        if value_type is str:
            command = CallCommand(service=value)
        elif value_type is dict:
            command = CallCommand(**case.snakify_dict(value))
        else:
            raise TypeError(
                'incompatible type "{}" for CallCommand value'.format(
                    value_type))
    else:
        raise KeyError('unknown command key "{}"'.format(key))

    return command


class Service:
    def __init__(self, name: str, is_entrypoint: bool, response_size: int,
                 script: List[Command]) -> None:
        self.name = name
        self.is_entrypoint = is_entrypoint
        self.response_size = response_size
        self.script = script


def _service_from_dict(d: Dict[str, Any]) -> Service:
    commands = []  # type: List[Command]
    if 'script' in d:
        command_dicts = d['script']
        commands = [_command_from_dict(d) for d in command_dicts]
    return Service(script=commands, **d)


class Topology:
    def __init__(self, services: List[Service]) -> None:
        self.services = services


def from_yaml(yaml_str: str) -> Topology:
    """Returns a Topology from yaml_str."""
    d = yaml.safe_load(yaml_str)
    snake_case_dict = case.snakify_dict(d)
    logging.debug('YAML as Python dict: %s', snake_case_dict)
    print('YAML as Python dict: %s', snake_case_dict)
    return _from_dict(snake_case_dict)


def _from_dict(d: Dict[str, Any]) -> Topology:
    """Returns a Topology from d. Keys should be in snake_case."""
    service_dicts = d['services']
    services = [_service_from_dict(d) for d in service_dicts]
    return Topology(services=services)

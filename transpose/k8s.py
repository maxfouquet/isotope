import json
from typing import List

import kubernetes
import yaml

import topology


def to_json(yaml_str: str) -> str:
    d = yaml.safe_load(yaml_str)
    return json.dumps(d)


def to_topology(json_str: str) -> topology.Topology:
    json.loads(json_str, object_hook=lambda d: topology.Topology(**d))


def topology_to_manifests(topology: topology.Topology) -> str:
    pass

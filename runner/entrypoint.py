import yaml

from . import consts


def extract_url(topology_path: str) -> str:
    """Returns the in-cluster URL to access the service graph's entrypoint."""
    with open(topology_path, 'r') as f:
        topology = yaml.load(f)

    services = topology['services']
    entrypoint_services = [svc for svc in services if svc.get('isEntrypoint')]
    if len(entrypoint_services) != 1:
        raise ValueError(
            'topology at {} should only have one entrypoint'.format(
                topology_path))
    entrypoint_name = entrypoint_services[0]['name']
    url = 'http://{}.{}.svc.cluster.local:{}'.format(
        entrypoint_name, consts.SERVICE_GRAPH_NAMESPACE, consts.SERVICE_PORT)
    return url

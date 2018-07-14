import os

_RESOURCES_DIR = os.path.realpath(
    os.path.join(os.getcwd(), os.path.dirname(__file__)))

STACKDRIVER_PROMETHEUS_GEN_YAML_PATH = os.path.join(
    _RESOURCES_DIR, 'stackdriver-prometheus.gen.yaml')
SERVICE_GRAPH_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'service-graph.gen.yaml')
ISTIO_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR, 'istio.gen.yaml')
ISTIO_INGRESS_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                       'istio-ingress.gen.yaml')

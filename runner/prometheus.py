import logging
import textwrap
from typing import Dict, Iterable, Tuple

import jinja2

TEMPLATE = jinja2.Template(
    textwrap.dedent("""\
        serviceMonitors:
        - name: service-graph-monitor
          selector:
            matchLabels:
              app: service-graph
          namespaceSelector:
            matchNames:
            - service-graph
          endpoints:
          - targetPort: 8080
            metricRelabelings:
            {%- for key, value in labels.items() %}
            - targetLabel: "{{ key }}"
              replacement: "{{ value }}"
            {%- endfor %}
        - name: istio-mixer-monitor
          selector:
            matchLabels:
              istio: mixer
          namespaceSelector:
            matchNames:
            - istio-system
          endpoints:
          - targetPort: 42422
            metricRelabelings:
            {%- for key, value in labels.items() %}
            - targetLabel: "{{ key }}"
              replacement: "{{ value }}"
            {%- endfor %}
        storageSpec:
          volumeClaimTemplate:
            spec:
              # It's necessary to specify "" as the storageClassName
              # so that the default storage class won't be used, see
              # https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1
              storageClassName: ""
              volumeName: prometheus-persistent-volume
              accessModes:
              - ReadWriteOnce
              resources:
                requests:
                  storage: 10G
    """))


def values_yaml(labels: Dict[str, str]) -> str:
    logging.info('generating Prometheus configuration')
    return TEMPLATE.render(labels=labels)

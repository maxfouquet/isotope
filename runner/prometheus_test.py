import collections
import textwrap

import yaml

from . import prometheus


def test_values_should_return_correct_yaml():
    expected_yaml = textwrap.dedent("""\
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
            - targetLabel: "user"
              replacement: "tjberry"
            - targetLabel: "custom"
              replacement: "stuff"
        - name: client-monitor
          selector:
            matchLabels:
              app: client
          namespaceSelector:
            matchNames:
            - default
          endpoints:
          - targetPort: 8080
            metricRelabelings:
            - targetLabel: "user"
              replacement: "tjberry"
            - targetLabel: "custom"
              replacement: "stuff"
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
            - targetLabel: "user"
              replacement: "tjberry"
            - targetLabel: "custom"
              replacement: "stuff"
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
    """)
    expected = yaml.load(expected_yaml)

    # Use an OrderedDict to prevent flakiness from iterating dictionaries.
    labels = collections.OrderedDict([
        ('user', 'tjberry'),
        ('custom', 'stuff'),
    ])

    actual_yaml = prometheus.values_yaml(labels)
    actual = yaml.load(actual_yaml)

    assert expected == actual

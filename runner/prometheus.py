import logging
import textwrap
from typing import Any, Dict, List

import yaml

from . import consts


def values_yaml(labels: Dict[str, str]) -> str:
    """Returns Prometheus Helm values with relabelings to include labels."""
    logging.info('generating Prometheus configuration')
    config = _get_config(labels)
    return yaml.dump(config, default_flow_style=False)


def _get_config(labels: Dict[str, str]) -> Dict[str, Any]:
    return {
        'deployAlertManager': False,
        'deployExporterNode': False,
        'deployGrafana': False,
        'deployKubeControllerManager': False,
        'deployKubeDNS': False,
        'deployKubeEtcd': False,
        'deployKubelets': False,
        'deployKubeScheduler': False,
        'deployKubeState': False,
        'prometheus': _get_prometheus_config(labels)
    }


def _get_prometheus_config(labels: Dict[str, str]) -> Dict[str, Any]:
    metric_relabelings = _get_metric_relabelings(labels)
    return {
        'serviceMonitors': [
            _get_service_monitor('service-graph-monitor', 8080,
                                 consts.SERVICE_GRAPH_NAMESPACE,
                                 {'app': 'service-graph'}, metric_relabelings),
            _get_service_monitor('client-monitor', 42422,
                                 consts.DEFAULT_NAMESPACE, {'app': 'client'},
                                 metric_relabelings),
            _get_service_monitor('istio-mixer-monitor', 42422,
                                 consts.ISTIO_NAMESPACE, {'istio': 'mixer'},
                                 metric_relabelings),
        ],
        'storageSpec':
        _get_storage_spec(),
    }


def _get_service_monitor(
        name: str, port: int, namespace: str, match_labels: Dict[str, str],
        metric_relabelings: List[Dict[str, Any]]) -> Dict[str, Any]:
    return {
        'name':
        name,
        'endpoints': [{
            'targetPort': port,
            'metricRelabelings': metric_relabelings,
        }],
        'namespaceSelector': {
            'matchNames': [namespace],
        },
        'selector': {
            'matchLabels': match_labels,
        },
    }


def _get_metric_relabelings(labels: Dict[str, str]) -> List[Dict[str, Any]]:
    return [{
        'targetLabel': key,
        'replacement': value,
    } for key, value in labels.items()]


def _get_storage_spec() -> Dict[str, Any]:
    return {
        'volumeClaimTemplate': {
            'spec': {
                'accessModes': ['ReadWriteOnce'],
                'resources': {
                    'requests': {
                        'storage': '10G',
                    },
                },
                'volumeName': 'prometheus-persistent-volume',
                'storageClassName': '',
            },
        },
    }

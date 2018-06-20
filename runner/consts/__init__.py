import datetime
import os

DEFAULT_NAMESPACE = 'default'
MONITORING_NAMESPACE = 'monitoring'
ISTIO_NAMESPACE = 'istio-system'
SERVICE_GRAPH_NAMESPACE = 'service-graph'

CLIENT_JOB_NAME = 'client'
SERVICE_GRAPH_SERVICE_SELECTOR = 'role=service'

_GOPATH_ENV = os.getenv('GOPATH')
_HOME_ENV = os.getenv('HOME')
_HOME = os.path.join('/', 'tmp') if _HOME_ENV is None else _HOME_ENV
GOPATH = os.path.join(_HOME, 'go') if _GOPATH_ENV is None else _GOPATH_ENV
ISTIO_REPO_PATH = os.path.join(GOPATH, 'istio.io', 'istio')

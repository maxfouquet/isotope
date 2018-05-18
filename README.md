# Service Grapher

## service-graph.yaml

Describes a service graph to be tested which mocks a real world service-oriented
architecture.

### Full example

```yaml
apiVersion: v1alpha1
kind: MockServiceGraph
defaults:
  type: grpc
  requestSize: 1 KB
  responseSize: 16 KB
services:
- name: a
  errorRate: 0.01%
  script:
  - sleep: 100ms
- name: b
  type: grpc
- name: c
  script:
  - call:
      service: a
      size: 10K
  - call: b
- name: d
  script:
  - - call: a
    - call: c
  - sleep: 10ms
  - call: b
```

Represents a service graph like:

![service-graph](graph.png)

Generates a Kubernetes manifest like:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scripts
data:
  a: |
    errorRate: 0.0001
    name: a
    responseSize: 16KiB
    script:
    - sleep: 100ms
    type: http
  b: |
    name: b
    responseSize: 16KiB
    type: grpc
  c: |
    name: c
    responseSize: 16KiB
    script:
    - call:
        service: a
        size: 10KiB
    - call:
        service: b
        size: 1KiB
    type: http
  d: |
    name: d
    responseSize: 16KiB
    script:
    - - call:
          service: a
          size: 1KiB
      - call:
          service: c
          size: 1KiB
    - sleep: 10ms
    - call:
        service: b
        size: 1KiB
    type: http
---
apiVersion: v1
kind: Service
metadata:
  name: a
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a
spec:
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
---
apiVersion: v1
kind: Service
metadata:
  name: b
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b
spec:
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
---
apiVersion: v1
kind: Service
metadata:
  name: c
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: c
spec:
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
---
apiVersion: v1
kind: Service
metadata:
  name: d
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d
spec:
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
---
```

### Specification

```yaml
apiVersion: {{ Version }} # Required. K8s-like API version.
kind: MockServiceGraph
default: # Optional. Default to empty map.
  type: {{ "http" | "grpc" }} # Optional. Default "http".
  errorRate: {{ Percentage }} # Optional. Default 0%.
  requestSize: {{ ByteSize }} # Optional. Default 0.
  responseSize: {{ ByteSize }} # Optional. Default 0.
  script: {{ Script }} # Optional. See below for spec.
services: # Required. List of services in the graph.
- name: {{ ServiceName }}: # Required. Name of the service.
  type: {{ "http" | "grpc" }} # Optional. Default "http".
  responseSize: {{ ByteSize }} # Optional. Default 0.
  errorRate: {{ Percentage }} # Optional. Overrides default.
  script: {{ Script }} # Optional. See below for spec.
```

#### Default

At the global scope a `default` map may be placed to indicate settings which
should hold for omitted settings for its current and nested scopes.

Default-able settings include `type`, `script`, `responseSize`,
`requestSize`, and `errorRate`.

##### Example

```yaml
apiVersion: v1alpha1
default:
  errorRate: 0.1%
  requestSize: 100KB
  # responseSize: 0 # Inherited from default.
  # type: "http" # Inherited from default.
  # script: [] # Inherited from default (acts like an echo server).
services:
- name: a
  memoryUsage: 80%
  script:
  - call: b # payloadSize: 100KB # Inherited from default.
  - call:
      service: b
      payloadSize: 80B
  # computeUsage: 10% # Inherited from default.
  # errorRate: 10% # Inherited from default.
- name: b
  errorRate: 5%
  # computeUsage: 10% # Inherited from default.
  # memoryUsage: 0% # Inherited from default.
  # script: [] # Inherited from default.
```

#### Script

`script` is a list of high level steps which run when the service is called.

Each step is executed sequentially and may contain either a single command or
a list of commands. If the step is a list of commands, each command in that
sub-list is executed concurrently (this effect is not recursive; there may
only be one level of nested lists).

The script is always _started when the service is called_ and _ends by
responding to the calling service_.

##### Commands

Each step in the script includes a command.

###### Sleep

`sleep`: Pauses for a duration. Useful for simulating processing time.

```yaml
sleep: {{ Duration }}
```

###### Send Request

`call`: Sends a HTTP/gRPC request (depending on the receiving service's type)
to another service.

```yaml
call: {{ ServiceName }}
```

OR

```yaml
call:
  service: {{ ServiceName }}
  payloadSize: {{ ByteSize (e.g. 1 KB) }}
```

##### Examples

Call A, then call B _sequentially_:

```yaml
script:
- call: A
- call: B
```

Call A, B, and C _concurrently_, sleep to simulate work, and finally call D:

```yaml
script:
- - call: A
  - call: B
  - call: C
- sleep: 10ms
- call: D
```

## Architecture and pipeline

### 1. Parsing

A .yaml file is converted into a service graph data structure which is then
output as a Kubernetes manifest. However, another tool could be used to
generate Consul artifacts instead, for example.

### 2. Setup

A testing cluster is preexistent, with the desired "level" of Istio already
installed. Therefore, the manifest only needs to be applied to the cluster via
`kubectl apply`.

Each pod in the manifest will be running the same container image (except the
Istio images). A client will send requests to each service with their
instructions for the duration of the test (TODO: unless there is some way to
include data in the manifest, then this could be done at the manifest level).

### 3. Test

Should send requests to every leaf in the tree (i.e. C, D).

Each node should be monitored by its CPU and memory usage and by latencies of
each request.

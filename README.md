# Service Grapher

## service-graph.yaml

Describes a service graph to be tested which mocks a real world service-oriented
architecture.

### Full example

```yaml
apiVersion: v1alpha1
services:
  A:
    computeUsage: 1%
    memoryUsage: 20%
    errorRate: 0.01%
    script:
    - sleep: 100ms
  B:
    computeUsage: 10%
    memoryUsage: 10%
  C:
    script:
    - get: A
    - post: B
  D:
    # Call A and C concurrently, process, then call B.
    script:
    - - get: A
      - get: C
    - sleep: 10ms
    - delete: B
```

Represents a service graph like:

![service-graph](graph.png)

Generates a Kubernetes manifest like: (TODO: Pretty sure I'm using ConfigMaps
incorrectly).

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: a-script
data:
  script:
  - sleep: 100ms
---
apiVersion: v1
kind: Service
metadata:
  name: A
spec:
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: A
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
        ports:
        - containerPort: 80
        env:
        - name: SCRIPT
          valueFrom:
            configMapKeyRef:
              name: a-script
              key: A-SCRIPT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: b-script
data:
  script:
---
apiVersion: v1
kind: Service
metadata:
  name: B
spec:
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: B
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
        ports:
        - containerPort: 80
        env:
        - name: SCRIPT
          valueFrom:
            configMapKeyRef:
              name: b-script
              key: B-SCRIPT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: c-script
data:
  script:
  - call: A
  - call: B
---
apiVersion: v1
kind: Service
metadata:
  name: C
spec:
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: C
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
        ports:
        - containerPort: 80
        env:
        - name: SCRIPT
          valueFrom:
            configMapKeyRef:
              name: c-script
              key: C-SCRIPT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: d-script
data:
  script:
  - - call: A
    - call: C
  - sleep: 10ms
  - call: B
---
apiVersion: v1
kind: Service
metadata:
  name: D
spec:
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: D
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: performance-test
        image: istio.gcr.io/performance-test
        ports:
        - containerPort: 80
        env:
        - name: SCRIPT
          valueFrom:
            configMapKeyRef:
              name: d-script
              key: D-SCRIPT
---
```

### Specification

```yaml
apiVersion: {{ Version }} # Required. K8s-like API version.
services: # Required. List of services in the graph.
  default: # Optional. Sets the inherited defaults of other services. Default to empty map.
    computeUsage: {{ Percentage }} # Optional. Default 0%.
    memoryUsage: {{ Percentage }} # Optional. Default 0%.
    errorRate: {{ Percentage }} # Optional. Default 0%.
    script: {{ Script }} # Optional. Default echo server. See below for spec.
  {{ ServiceName }}: # Required. Name of the service.
    {{ overrided }}
```

#### Script

`script` is a list of high level steps which run when the service is called.

Each step is executed sequentially and may contain either a single command or
a list of commands. If the step is a list of commands, each command in that
sub-list is executed concurrently (this effect is not recursive; there may
only be one level of nested lists).

The script is always started when the service is called and ends by
responding to the calling service.

##### Commands

Each step in the script includes a command.

`sleep`: Pauses for a duration. Useful for simulating processing time.

```yaml
sleep: {{ Duration }}
```

`get`, `head`, `post`, `put`, `delete`, `connect`, `options`, `trace`,
`patch`: Sends the respective HTTP request to another service.

```yaml
{{ HttpMethod }}: {{ ServiceName }}
```

##### Examples

Get B, then post to B _sequentially_:

```yaml
script:
- get: B
- post: B
```

GET A, B, and C _concurrently_, sleep to simulate work, and finally POST to D:

```yaml
script:
- - get: A
  - get: B
  - get: C
- sleep: 10ms
- post: D
```

## Architecture and pipeline

### 1. Parsing

A .yaml file is converted into a service graph data structure which is then
output as a Kubernetes manifest. However, another tool could be used to
generate Consul artifacts instead, for example.

### 2. Setup

The manifest is applied to a testing cluster with automatic scaling (TODO:
Will this be sufficient?).

Istio is then injected at the configured level (TODO: Or is this covered at
the manifest level?).

Each pod in the manifest will be running the same container image (except the
Istio images). A client will send requests to each service with their
instructions for the duration of the test (TODO: unless there is some way to
include data in the manifest, then this could be done at the manifest level).

### 3. Test

Should send requests to every leaf in the tree (i.e. C, D).

Each node should be monitored by its CPU and memory usage and by latencies of
each request.

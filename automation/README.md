# Automation

This subdirectory contains the code for automating topology tests.

## Contents

- `convert` contains the Go code for converting topology YAML to other formats
- `runner` contains the Python module run by `run_tests.py`, which executes:

  ```txt
  read configuration
  create cluster
  add prometheus
  for each topology:
    convert topology to Kubernetes YAML
    for each environment (none, istio, sidecars only, etc.):
      update Prometheus labels
      deploy environment
      deploy topology
      run load test
      delete topology
      delete environment
  delete cluster
  ```

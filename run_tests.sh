#!/bin/bash
#
# Run a test on GKE.
# - Operates on the current kubectl configuration and requires input.yaml to be
#   present in the working directory
# - Assumes GKE cluster is already present with an operational Prometheus
# - Runs two tests for each argument:
#   - No Istio
#   - With Istio installed via the Helm template
#
# TODO: Convert this to native Go or Python code (see branch native-testing).

create_cluster_admin_binding() {
  kubectl create clusterrolebinding cluster-admin-binding \
      --clusterrole cluster-admin \
      --user "$(gcloud config get-value account)"
}

log() {
  printf '> %s\n' "$*"
}

error() {
  echo "$*" >&2
}

suppress() {
  {
    "$@"
  } &> /dev/null
}

wait_until() {
  until "$@"; do
    sleep 1
  done
}

wait_for_deployments() {
  local namespace="${1}"
  local deployments
  deployments=$(kubectl --namespace "${namespace}" get deployments \
      -o jsonpath='{.items[*].metadata.name}')
  log "waiting for all deployments in ${namespace} (${deployments}) to rollout"
  for deployment in ${deployments}; do
    suppress kubectl --namespace istio-system \
        rollout status "deployment/${deployment}"
  done
}

namespace_is_deleted() {
  local namespace="$1"
  ! suppress kubectl get namespace "${namespace}"
}

gen_yaml() {
  local path="$1"
  log "generating manifests for Kubernetes from ${path}"
  go run main.go performance kubernetes "${path}"
}

service_graph_is_ready() {
  local statuses
  statuses=$(kubectl --namespace service-graph get pods \
      --selector role=service \
      -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}')
  ! ([ -z "${statuses}" ] || echo "${statuses}" | suppress grep False)
}

create_service_graph() {
  log "creating service graph"
  kubectl create namespace service-graph || return
  # Does nothing if Istio is not installed.
  kubectl label namespace service-graph istio-injection=enabled || return
  kubectl create -f service-graph.yaml || return

  wait_for_deployments service-graph
  # TODO: Why is this extra buffer necessary?
  sleep 30
}

delete_service_graph() {
  log "deleting service graph"
  kubectl delete -f service-graph.yaml
  kubectl delete namespace service-graph
  wait_until namespace_is_deleted service-graph
}

client_job_is_complete() {
  local status
  status=$(kubectl get job client \
      -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
  echo "${status}" | suppress grep True
}

create_client_job() {
  log "creating client job"
  kubectl create -f client.yaml || return

  log "waiting for client job to complete"
  wait_until client_job_is_complete || return
}

save_client_job_logs() {
  local path="$1"
  log "fetching client job logs to ${path}"
  kubectl logs job/client > "${path}"
}

delete_client_job() {
  log "deleting client job"
  kubectl delete -f client.yaml
}

create_istio() {
  log "creating Istio components"
  kubectl create namespace istio-system || return
  kubectl create -f istio.yaml || return
  wait_for_deployments istio-system || return
}

delete_istio() {
  log "deleting Istio components"
  # Must delete from istio.yaml to remove the CRDs.
  kubectl delete -f istio.yaml
  kubectl delete namespace istio-system
  wait_until namespace_is_deleted istio-system
}

test_service_graph() {
  local path="$1"

  create_service_graph || return
  create_client_job || return
  save_client_job_logs "${path}"
  delete_client_job
  delete_service_graph
}

test_service_graph_with_istio() {
  local path="$1"

  create_istio || return
  test_service_graph "${path}" || return
  delete_istio
}

main() {
  for input_path in "$@"; do
    local base_name
    base_name=$(basename -- "${input_path}")
    local file_name_no_ext="${base_name%.*}"

    gen_yaml "${input_path}" || return

    test_service_graph "${file_name_no_ext}-no-istio.log" || return

    test_service_graph_with_istio "${file_name_no_ext}-full-istio.log" || return
  done
}

main "$@"

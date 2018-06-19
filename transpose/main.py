#!/usr/bin/env python3

from kubernetes import client
import yaml


def main():
  api = client.ApiClient()
  obj = create_deployment_object()
  obj2 = api.sanitize_for_serialization(obj)
  print(yaml.dump(obj))
  print("---")
  print(yaml.dump(obj2))


def create_deployment_object():
  # Configureate Pod template container
  container = client.V1Container(
      name="nginx",
      image="nginx:1.7.9",
      ports=[client.V1ContainerPort(container_port=80)])
  # Create and configurate a spec section
  template = client.V1PodTemplateSpec(
      metadata=client.V1ObjectMeta(labels={"app": "nginx"}),
      spec=client.V1PodSpec(containers=[container]))
  # Create the specification of deployment
  spec = client.ExtensionsV1beta1DeploymentSpec(replicas=3, template=template)
  # Instantiate the deployment object
  deployment = client.ExtensionsV1beta1Deployment(
      api_version="extensions/v1beta1",
      kind="Deployment",
      metadata=client.V1ObjectMeta(name="some-name"),
      spec=spec)

  return deployment


if __name__ == "__main__":
  main()

# Runs `run_tests.py` under Prow with kubectl pointing to a Boskos-provisioned
# cluster.
#
# Dependencies:
# - python3
# - kubectl
# - helm

FROM tahler/convert:0.1 AS convert
FROM lachlanevenson/k8s-kubectl:v1.11.1 AS kubectl
FROM linkyard/docker-helm:2.9.1 AS helm

FROM ubuntu:18.04

COPY --from=convert /usr/local/bin/convert /usr/local/bin/convert

COPY --from=kubectl /usr/local/bin/kubectl /usr/local/bin/kubectl

COPY --from=helm /bin/helm /usr/local/bin/helm

RUN apt-get update

RUN apt-get install -y python3
RUN apt-get install -y python3-pip

WORKDIR /isotope/automation
COPY . .

RUN pip3 install -r runner/requirements.txt

ENTRYPOINT [ "./run_tests.py", "isotope.toml" ]

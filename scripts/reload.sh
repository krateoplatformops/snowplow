#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

kubectl delete -f manifests/deploy.bff.yaml

${SCRIPT_DIR}/build.sh

kubectl apply -f manifests/deploy.bff.yaml


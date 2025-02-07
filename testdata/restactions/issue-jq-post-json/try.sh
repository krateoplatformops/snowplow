#!/bin/bash

kubectl create namespace fireworksapp-system || true
kubectl create namespace  krateo-system || true
kubectl apply -f fireworksapps.composition.yaml
kubectl apply -f fireworksapps.composition.crd.yaml
kubectl apply -f yamlviewer.yaml

curl -v -G GET \
  -H 'x-krateo-user: cyberjoker' \
  -H 'x-krateo-groups: devs' \
  -d 'apiVersion=templates.krateo.io/v1' \
  -d 'resource=restactions' \
  -d 'namespace=fireworksapp-system' \
  -d 'name=composition-tabpane-yamlviewer-row-column-1-panel-yamlviewer' \
  "http://127.0.0.1:30081/call"
#!/bin/bash

kubectl create namespace demo-system || true

kubectl apply -f crds/templates.krateo.io_restactions.yaml

kubectl apply -f testdata/widgets/button.crd.yaml
kubectl apply -f testdata/rbac.widgets.yaml
kubectl apply -f testdata/widgets/button.sample.yaml

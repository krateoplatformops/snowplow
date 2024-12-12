#!/bin/bash

kubectl apply -f crds/templates.krateo.io_customforms.yaml
kubectl apply -f testdata/customforms/fireworksapp-crd.yaml
kubectl apply -f testdata/customforms/rbac.yaml
kubectl apply -f testdata/customforms/sample.yaml

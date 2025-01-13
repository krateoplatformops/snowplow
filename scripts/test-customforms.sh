#!/bin/bash

kubectl apply -f crds/templates.krateo.io_customforms.yaml
kubectl apply -f testdata/customforms

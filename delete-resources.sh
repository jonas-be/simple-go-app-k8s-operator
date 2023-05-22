#!/bin/bash

kubectl delete simplegoappdeployments.simple-go-app-k8s-operator.jonasbe.de simplegoappdeployment-sample

# Remove finalizers
kubectl patch deployments.apps simple-go-app-backend -p '{"metadata":{"finalizers":null}}' --type=merge

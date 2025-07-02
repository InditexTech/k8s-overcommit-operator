#!/bin/bash

# SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
# SPDX-FileContributor: enriqueavi@inditex.com
#
# SPDX-License-Identifier: Apache-2.0

set -e
echo "Mounting image in the Kind cluster"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.17.0/cert-manager.yaml
kubectl wait --for=condition=available --timeout=40s deployment/cert-manager-webhook -n cert-manager
make docker-build IMG=k8s-overcommit/webhook:teste2e
kind load docker-image k8s-overcommit/webhook:teste2e --name kuttl-cluster
echo "Mounted image in the Kind cluster, instaling operator"
helm install k8s-overcommit chart --set createClasses=false --set createNamespace=true --set namespace=k8s-overcommit --set deployment.image.tag=teste2e --set deployment.image.registry=docker.io --set deployment.image.image=k8s-overcommit/webhook
echo "Operator installed, waiting for the deployment to be ready"
kubectl wait --for=condition=available --timeout=50s deployment/k8s-overcommit-operator -n $(yq eval '.namespace' chart/values.yaml)
echo "Operator deployment ready, installing the overcommit CR"
kubectl apply -f hack/overcommitTest.yaml
sleep 5
kubectl wait --for=condition=available --timeout=40s deployment/k8s-overcommit-overcommitclass-controller  -n $(yq eval '.namespace' chart/values.yaml)
echo "Operator ready, installing the overcommit class"
sleep 10
kubectl apply -f hack/overcommitClassTests.yaml
sleep 10
kubectl wait --for=condition=available --timeout=40s deployment/test-overcommit-webhook -n $(yq eval '.namespace' chart/values.yaml)
echo "tests environment ready"

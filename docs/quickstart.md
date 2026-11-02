<!--
SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
SPDX-FileContributor: enriqueavi@inditex.com

SPDX-License-Identifier: CC-BY-4.0
-->

# 🚀 Quick Start

> [!IMPORTANT]
> **Prerequisites**: Ensure **cert-manager** is installed in your cluster before deploying the operator.

Choose your preferred installation method:

## 📦 Installation Methods

### 🎯 Method 1: Helm Installation (Recommended)

#### 1️⃣ Clone the Repository

Clone the repository to your local machine:

```bash
git clone https://github.com/InditexTech/k8s-overcommit-operator.git
cd k8s-overcommit-operator
```

#### 2️⃣ Configure Values

Edit the [`values.yaml`](../chart/values.yaml) file to customize your deployment. Below is an example configuration:

```yaml
# Example configuration
deployment:
  image:
    registry: ghcr.io
    image: inditextech/k8s-overcommit-operator
    tag: 1.0.0
```

#### 3️⃣ Install with Helm

Install the operator using Helm:

```bash
helm install k8s-overcommit-operator chart
```

### 🔧 Method 2: OLM Installation

#### 1️⃣ Install the CatalogSource

For OpenShift or clusters with OLM installed, apply the catalog source:

```bash
kubectl apply -f https://raw.githubusercontent.com/InditexTech/k8s-overcommit-operator/refs/heads/main/deploy/catalog_source.yaml
```

#### 2️⃣ Apply the OperatorGroup

Apply the operator group configuration:

```bash
kubectl apply -f https://raw.githubusercontent.com/InditexTech/k8s-overcommit-operator/refs/heads/main/deploy/operator_group.yaml
```

#### 3️⃣ Create the Subscription (Alternative)

You can create your own subscription or use the default [`subscription.yaml`](../deploy/subscription.yaml). Below is an example:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: k8s-overcommit-operator
  namespace: operators
spec:
  channel: alpha
  name: k8s-overcommit-operator
  source: community-operators
  sourceNamespace: olm
```

Apply the subscription:

```bash
kubectl apply -f https://raw.githubusercontent.com/InditexTech/k8s-overcommit-operator/refs/heads/main/deploy/subscription.yaml
```

#### 4️⃣ Validation

After installation, validate that the operator is running:

```bash
kubectl get pods -n k8s-overcommit
```

---

## 🎯 Create the Custom Resources

The operator is configured through two cluster-scoped custom resources from the
`overcommit.inditex.dev/v1alphav1` API group (see
[`api/v1alphav1/`](../api/v1alphav1/)). They are **not** created by the Helm chart by default
(`createOvercommit: false`, `createClasses: false`), so create them once the operator is running.

### 1️⃣ Create the `Overcommit` resource

`Overcommit` is a **singleton**: the CRD rejects any object whose `metadata.name` is not `cluster`.
It tells the operator which label selects an `OvercommitClass` on pods and namespaces.

```yaml
apiVersion: overcommit.inditex.dev/v1alphav1
kind: Overcommit
metadata:
  name: cluster
spec:
  # Required (min length 1). Default of the Helm chart: inditex.com/overcommit-class
  overcommitLabel: inditex.com/overcommit-class
  # Optional: labels/annotations added to the generated resources
  labels:
    example: "label"
  annotations:
    example: "annotation"
  # Optional: scheduling constraints for the generated deployments
  nodeSelector: {}
  tolerations: []
```

### 2️⃣ Create one or more `OvercommitClass` resources

```yaml
apiVersion: overcommit.inditex.dev/v1alphav1
kind: OvercommitClass
metadata:
  name: my-class
spec:
  cpuOvercommit: 0.1
  memoryOvercommit: 0.8
  excludedNamespaces: ".*(^(openshift|k8s-overcommit|kube).*).*"
  isDefault: true
```

A complete example, including the optional fields, is available in
[`overcommitClass-example.yaml`](./overcommitClass-example.yaml).

### 3️⃣ Label workloads

```bash
kubectl label namespace my-namespace inditex.com/overcommit-class=my-class
# or, for a single workload:
kubectl label deployment my-app inditex.com/overcommit-class=my-class
```

Pods whose namespace or pod template carries the `overcommitLabel` key get their CPU and memory
**requests** set to `limits × overcommit ratio`; limits are never modified.

### 4️⃣ Verify

```bash
kubectl get overcommit cluster                 # Overcommit singleton
kubectl get overcommitclass                    # columns: CPU, Memory, Default (short names: oc, ocs)
kubectl get overcommitclass my-class -o yaml   # status.conditions / status.resources
```

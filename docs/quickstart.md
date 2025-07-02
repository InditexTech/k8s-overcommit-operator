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
apiVersion: operators.coreos.com/v1alphav1
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

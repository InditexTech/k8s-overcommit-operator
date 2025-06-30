<!--
SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
SPDX-FileContributor: enriqueavi@inditex.com

SPDX-License-Identifier: CC-BY-4.0
-->

# 🚀 Quick Start

> [!IMPORTANT]
> **Prerequisites**: You need to have **cert-manager** installed in your cluster before deploying the operator.

Choose your preferred installation method:

### 📦 Installation Methods

<details>
<summary><strong>🎯 Method 1: Helm Installation (Recommended)</strong></summary>

#### 1️⃣ Clone the Repository

```bash
git clone https://github.com/InditexTech/k8s-overcommit-operator.git
cd k8s-overcommit-operator
```

#### 2️⃣ Configure Values

Edit the [`chart/values.yaml`](chart/values.yaml) file to customize your deployment:

```yaml
# Example configuration
deployment:
  image:
    registry: ghcr.io
    image: inditextech/k8s-overcommit-operator
    tag: 1.0.0
```

#### 3️⃣ Install with Helm

```bash
helm install k8s-overcommit-operator chart
```

</details>

<details>
<summary><strong>🔧 Method 2: OLM Installation</strong></summary>

#### 1️⃣ Install the catalog source

For OpenShift or clusters with OLM installed:

```bash
kubectl apply -f https://github.com/InditexTech/k8s-overcommit-operator/deploy/catalog_source.yaml
```

#### 2️⃣ Apply the operatorGroup

```bash
kubectl apply -f https://github.com/InditexTech/k8s-overcommit-operator/deploy/operator_group.yaml
```

#### 3️⃣ Create the Subscription (Alternative)

You can create yot own or use the one in the route *https://github.com/InditexTech/k8s-overcommit-operator/deploy/subsciption.yaml*

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

</details>

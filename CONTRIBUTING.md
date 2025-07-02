<!--
SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
SPDX-FileContributor: enriqueavi@inditex.com

SPDX-License-Identifier: CC-BY-4.0
-->

# Contributing

Thank you for your interest in contributing to this project! We value and appreciate any contributions you can make.
To maintain a collaborative and respectful environment, please consider the following guidelines when contributing to
this project.

## Prerequisites

- Before starting to contribute to the code, you must first sign the
  [Contributor License Agreement (CLA)](https://github.com/InditexTech/foss/blob/main/documents/CLA.pdf).
  Detailed instructions on how to proceed can be found [here](https://github.com/InditexTech/foss/blob/main/CONTRIBUTING.md).

## How to Contribute

1. Open an issue to discuss and gather feedback on the feature or fix you wish to address.
2. Fork the repository and clone it to your local machine.
3. Create a new branch to work on your contribution: `git checkout -b your-branch-name`.
4. Make the necessary changes in your local branch.
5. Ensure that your code follows the established project style and formatting guidelines.
6. Perform testing to ensure your changes do not introduce errors.
7. Make clear and descriptive commits that explain your changes.
8. Push your branch to the remote repository: `git push origin your-branch-name`.
9. Open a pull request describing your changes and linking the corresponding issue.
10. Await comments and discussions on your pull request. Make any necessary modifications based on the received feedback.
11. Once your pull request is approved, your contribution will be merged into the main branch.

## Contribution Guidelines

- All contributors are expected to follow the project's [code of conduct](CODE_of_CONDUCT.md). Please be respectful and
considerate towards other contributors.
- Before starting work on a new feature or fix, check existing [issues](../../issues) and [pull requests](../../pulls)
to avoid duplications and unnecessary discussions.
- If you wish to work on an existing issue, comment on the issue to inform other contributors that you are working on it.
This will help coordinate efforts and prevent conflicts.
- It is always advisable to discuss and gather feedback from the community before making significant changes to the
project's structure or architecture.
- Ensure a clean and organized commit history. Divide your changes into logical and descriptive commits. We recommend to use the [Conventional Commits Specification](https://www.conventionalcommits.org/en/v1.0.0/)
- Document any new changes or features you add. This will help other contributors and project users understand your work
and its purpose.
- Be sure to link the corresponding issue in your pull request to maintain proper tracking of contributions.
- Remember to add license and copyright information following the [REUSE Specification](https://reuse.software/spec/#copyright-and-licensing-information).

## Development

Make sure that you have:

- Read the rest of the [`CONTRIBUTING.md`](CONTRIBUTING.md) sections.
- Go installed
- [operator-sdk](https://sdk.operatorframework.io/docs/) installed
- [Kuttl](https://kuttl.dev/docs/) installed for the e2e tests
- A kubernetes cluster with admin access (kind, minikube, or real cluster)
- kubectl configured to access your cluster

### Setting up the Development Environment

1. **Clone the repository:**
   ```bash
   git clone https://github.com/InditexTech/k8s-overcommit-operator
   cd k8s-overcommit-operator
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Install CRDs into the cluster:**
   ```bash
   make install
   ```

4. **Run the operator locally:**
   ```bash
   make run
   ```

### Development Workflow

#### Creating New APIs

When adding new Custom Resource Definitions (CRDs):

```bash
# Create a new API
operator-sdk create api --group overcommit --version v1alphav1 --kind YourNewResource --resource --controller
```

#### Building and Testing

1. **Run unit tests:**
   ```bash
   make test
   ```

2. **Run e2e tests:**
    For this, read the [docs](./docs/e2e-test.md)

3. **Build the operator:**
   ```bash
   make build
   ```

4. **Build and push Docker image:**
   ```bash
   make docker-build docker-push IMG=<registry>/k8s-overcommit-operator:<tag>
   ```

#### Code Generation

After modifying API types, regenerate code:

```bash
make generate
make manifests
```

#### Local Development

1. **Deploy the operator in development mode:**
   ```bash
   make deploy IMG=<registry>/k8s-overcommit-operator:<tag>
   ```

2. **Undeploy when finished:**
   ```bash
   make undeploy
   ```

### Code Style and Standards

1. **Follow Go conventions:**
   - Use `make lint` for linting and `make lint-fix` for fix
   - Follow effective Go guidelines
   - Use meaningful variable and function names

2. **Controller patterns:**
   - Implement idempotent reconciliation
   - Handle errors gracefully
   - Use structured logging
   - Follow controller-runtime patterns

### Common Make Targets

- `make help` - Show available commands
- `make test` - Run unit tests
- `make build` - Build the binary
- `make run` - Run locally
- `make docker-build` - Build container image
- `make deploy` - Deploy to cluster
- `make undeploy` - Remove from cluster
- `make install` - Install CRDs
- `make uninstall` - Remove CRDs
- `make generate` - Generate code
- `make manifests` - Generate manifests

## Project Structure

```
k8s-overcommit-operator/
├── .github/                          # GitHub workflows and templates
│   └── workflows/                    # CI/CD pipelines
├── api/                              # Kubernetes API definitions
│   └── v1alphav1/                     # API version v1alphav1
│       ├── overcommitclass_types.go  # OvercommitClass CRD definition
│       ├── overcommitclass_webhook.go # Webhook implementation
│       ├── overcommitclass_webhook_test.go # Webhook tests
│       ├── groupversion_info.go      # Group version info
│       └── zz_generated.deepcopy.go  # Generated deep copy methods
├── config/                           # Kubernetes manifests and configuration
│   ├── crd/                         # Custom Resource Definitions
│   │   ├── bases/                   # Base CRD definitions
│   │   └── kustomization.yaml       # Kustomize configuration
│   ├── default/                     # Default deployment configuration
│   │   ├── kustomization.yaml       # Main kustomization file
│   │   ├── manager_auth_proxy_patch.yaml
│   │   └── manager_config_patch.yaml
│   ├── manager/                     # Controller manager configuration
│   │   ├── kustomization.yaml
│   │   └── manager.yaml             # Manager deployment
│   ├── prometheus/                  # Prometheus monitoring
│   │   ├── kustomization.yaml
│   │   └── monitor.yaml            # ServiceMonitor definition
│   ├── rbac/                       # Role-Based Access Control
│   │   ├── auth_proxy_client_clusterrole.yaml
│   │   ├── auth_proxy_role.yaml
│   │   ├── auth_proxy_role_binding.yaml
│   │   ├── auth_proxy_service.yaml
│   │   ├── kustomization.yaml
│   │   ├── leader_election_role.yaml
│   │   ├── leader_election_role_binding.yaml
│   │   ├── overcommitclass_editor_role.yaml
│   │   ├── overcommitclass_viewer_role.yaml
│   │   ├── role.yaml               # Main controller role
│   │   ├── role_binding.yaml       # Role binding
│   │   └── service_account.yaml    # Service account
│   ├── samples/                    # Sample Custom Resources
│   │   └── overcommit_v1alphav1_overcommitclass.yaml
│   └── webhook/                    # Webhook configuration
│       ├── kustomization.yaml
│       ├── manifests.yaml
│       └── service.yaml
├── docs/                           # Documentation
│   ├── e2e-test.md                # End-to-end testing guide
│   └── images/                    # Documentation images
├── internal/                      # Internal application code
│   ├── controller/                # Controllers implementation
│   │   └── overcommitclass/       # OvercommitClass controller
│   │       ├── overcommitclass_controller.go      # Main controller logic
│   │       ├── overcommitclass_controller_test.go # Controller tests
│   │       └── suite_test.go      # Test suite setup
│   └── utils/                     # Utility functions
│       ├── cleanup.go             # Cleanup utilities
│       ├── getOvercommit.go       # Overcommit calculation utilities
│       └── getOvercommitClass.go  # OvercommitClass retrieval utilities
├── pkg/                           # Public packages
│   └── overcommit/                # Overcommit calculation logic
│       ├── calculate_values_from_labels.go # Label-based calculations
│       └── overcommit.go          # Core overcommit functionality
├── test/                          # Test files
│   └── e2e/                       # End-to-end tests
│       ├── e2e_test.go           # E2E test implementation
│       └── e2e_suite_test.go     # E2E test suite
├── .gitignore                     # Git ignore rules
├── .golangci.yml                  # Go linting configuration
├── CODE_OF_CONDUCT.md             # Code of conduct
├── CONTRIBUTING.md                # Contributing guidelines
├── Dockerfile                     # Container image definition
├── LICENSE                        # Project license
├── Makefile                       # Build and development commands
├── PROJECT                        # Operator SDK project configuration
├── README.md                      # Project documentation
├── go.mod                         # Go module definition
└── go.sum                         # Go module checksums
```

### Directory Details

#### `/api/v1alphav1/`

Contains the Kubernetes API definitions for the v1alphav1 version:

- **overcommitclass_types.go**: Defines the OvercommitClass Custom Resource Definition (CRD) structure
- **overcommitclass_webhook.go**: Implements admission webhooks for validation and mutation
- **groupversion_info.go**: Contains API group and version information

#### `/config/`

Kubernetes manifests organized by component:

- **crd/**: Custom Resource Definitions that extend Kubernetes API
- **rbac/**: Role-Based Access Control configurations for security
- **manager/**: Controller manager deployment specifications
- **webhook/**: Admission webhook configurations
- **samples/**: Example Custom Resource instances

#### `/internal/`

Private application code not intended for external use:

- **controller/**: Contains the reconciliation logic for Custom Resources
- **utils/**: Shared utility functions for internal operations

#### `/pkg/`

Public packages that can be imported by external projects:

- **overcommit/**: Core business logic for overcommit calculations

#### `/test/`

Test files organized by test type:

- **e2e/**: End-to-end integration tests using real Kubernetes clusters

#### Root Files

- **Makefile**: Provides development and build automation commands
- **Dockerfile**: Defines the container image for the operator
- **PROJECT**: Operator SDK configuration file
- **go.mod/go.sum**: Go module dependency management

#### Getting Help

- Check existing [issues](../../issues) for similar problems
- Review the [operator-sdk documentation](https://sdk.operatorframework.io/docs/)
- Consult [controller-runtime documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)

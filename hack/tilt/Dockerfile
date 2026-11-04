# SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
# SPDX-FileContributor: enriqueavi@inditex.com
#
# SPDX-License-Identifier: Apache-2.0

# hack/tilt/Dockerfile

FROM gcr.io/distroless/static:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
WORKDIR /
COPY hack/tilt/bin/manager manager

USER 65532:65532

ENTRYPOINT ["/manager"]

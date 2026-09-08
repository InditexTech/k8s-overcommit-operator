# SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
# SPDX-FileContributor: enriqueavi@inditex.com
#
# SPDX-License-Identifier: Apache-2.0

# hack/tilt/Dockerfile

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY hack/tilt/bin/manager manager

USER 65532:65532

ENTRYPOINT ["/manager"]

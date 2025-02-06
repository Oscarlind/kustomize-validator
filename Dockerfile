FROM alpine:latest as dependencies
ARG HELM_VERSION=v3.17.0 \
    KUSTOMIZE_VERSION=v5.6.0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /workspace
ADD https://get.helm.sh/helm-${HELM_VERSION}-${GOOS}-${GOARCH}.tar.gz dest/helm.tar.gz
ADD https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_${GOOS}_${GOARCH}.tar.gz dest/kustomize.tar.gz
RUN tar -xvzf dest/helm.tar.gz && \
    tar -xvzf dest/kustomize.tar.gz && \
    mv ${GOOS}-${GOARCH}/helm /usr/local/bin/helm && \
    mv kustomize /usr/local/bin/kustomize

# build final image
FROM alpine:latest
ENV ENABLE_VERBOSITY=false \
    ENABLE_ERROR_ONLY=false \
    BASE_PATH=/
WORKDIR /
COPY kustomization-validator /kustomize-validator
COPY --from=dependencies /usr/local/bin/helm /usr/local/bin/helm
COPY --from=dependencies /usr/local/bin/kustomize /usr/local/bin/kustomize
USER 65532:65532

ENTRYPOINT ["/kustomize-validator", "${BASE_PATH}", "-v", "${ENABLE_VERBOSITY}", "-e", "${ENABLE_ERROR_ONLY}"]

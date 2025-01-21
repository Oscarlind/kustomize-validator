# Build the manager binary
FROM golang:1.23 as builder
ARG TARGETOS \
    TARGETARCH \
    HELM_VERSION
WORKDIR /workspace
ADD . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o kustomize-validator main.go

FROM alpine:latest as dependencies
ARG TARGETOS \
    TARGETARCH=linux_amd64 \
    HELM_VERSION=v3.17.0 \
    KUSTOMIZE_VERSION=v5.6.0
WORKDIR /workspace
COPY https://get.helm.sh/helm-${HELM_VERSION}-${TARGETARCH}.tar.gz dest
COPY https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_${TARGETARCH}.tar.gz dest
RUN tar -xvf dest/helm-${HELM_VERSION}-${TARGETARCH}.tar.gz && \
    tar -xvf dest/kustomize_${KUSTOMIZE_VERSION}_${TARGETARCH}.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/helm && \
    mv kustomize /usr/local/bin/kustomize

FROM gcr.io/distroless/static:nonroot
ENV ENABLE_VERBOSITY=false \
    ENABLE_ERROR_ONLY=false \
    BASE_PATH=/
WORKDIR /
COPY --from=builder /workspace/kustomize-validator .
COPY --from=dependencies /usr/local/bin/helm /usr/local/bin/helm
COPY --from=dependencies /usr/local/bin/kustomize /usr/local/bin/kustomize
USER 65532:65532

ENTRYPOINT ["/kustomize-validator", "${BASE_PATH}", "-v ${ENABLE_VERBOSITY}", "-e ${ENABLE_ERROR_ONLY}"]

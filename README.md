# Kustomization.yaml valdidator

[![Docker Repository on Quay](https://quay.io/repository/redhat-consulting-services/kustomize-validator/status "Docker Repository on Quay")](https://quay.io/repository/redhat-consulting-services/kustomize-validator)

A simple tool that will validate that all of the Kustomization files are buildable. This tool expects a base directory to start the recursive search from. It will then look for all the kustomization.yaml files and try to build them. If the build fails, it will print out the path to the failed kustomization.yaml file.

## Prerequisites

* Go 1.23 or later
* [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) v5.6.0 or later
* [Helm](https://helm.sh/docs/intro/install/) v3.16.4 or later

## Integration in CI

This tool can be integrated to validate kustomization stryctyres when pushing code to a git server, e.g. GitHub. It works well together with a linting tool, such as [yamllint](https://yamllint.readthedocs.io/en/stable/#).

To set it up with a runner, use the provided container image in your CI pipeline. For example, you can use the following snippet to orient yourself on how to set it up in GitHub Actions. This example assumes that your kustomization files are located in the `kustomize-files` directory at the root of your repository.:

```yaml
name: release

on:
  push:
    tags:
      - 'v*'

env:
  REGISTRY: quay.io
  IMAGE_REPOSITORY: quay.io/redhat-consulting-services/kustomize-validator

jobs:
  validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5

      - name: Execute Kustomize Validator
        run: docker run --rm -v ${{ github.workspace }}/kustomize-data:/data ${{ env.IMAGE_REPOSITORY }}:latest kustomize-validator /data -ev
```

> **NOTE:** This action is just an example. The implementation of the example above has not been tested

## Example

Assuming you have the following directory structure:

```txt
├── base
│   ├── logging
│   │   └── kustomization.yaml
│   ├── monitoring
│   │   └── kustomization.yaml
│   └── argocd
│       └── kustomization.yaml
└── overlays
    └── cluster-a
        ├── logging
        │   ├── kustomization.yaml
        │   └── patch-url.yaml
        ├── monitoring
        │   ├── kustomization.yaml
        │   └── patch-remote-write.yaml
        └── argocd
            ├── kustomization.yaml
            └── patch-url.yaml
```

In this example, you have a base directory with three kustomization.yaml files and an overlays directory with three more kustomization.yaml files. You can run the tool against the 'overlays' directory to validate the structure and buildability of the kustomization files as well as kubernetes manifests.

Run the tool from any location, and provide the base directory as an argument. The tool will then recursively search for all kustomization.yaml or kustomization.yml files and try to build them.

```sh
kustomize-validator ./overlays/cluster-a
```

Parameters:

```bash
A tool to validate Kustomization files

Usage:
  kustomize-validator [flags]

Flags:
  -c, --check strings   check for arbitrary validation in rendered kustomize output.
                        Use glob:pattern for glob matching, e.g., glob:PAT*_ME to match PAT123_ME
                        or use the regex match pattern regex:app-.* to match app-123.
                        If no prefix is provided, literal substring matching is used (default). (default [PATCH_ME,patch_me])
  -e, --error-only      whether we should only log errors
  -h, --help            help for kustomize-validator
  -t, --table           output resources in table format
  -v, --verbose         verbose output
```

### Example output

```bash
kustomize-validator ./_tests

Validating Kustomization files ./_tests
[ERROR]: Error while executing kustomize in path: _tests/app1, validation failed: found 'PATCH_ME' in line 23 for resource apps/v1/Deployment/<none>/my-app
[OK]: Successfully executed kustomize on _tests/app3
[ERROR]: Error while executing kustomize in path: _tests/app2, validation failed: found 'PATCH_ME' in line 23 for resource apps/v1/Deployment/<none>/my-app
Total:  3
Success:  1
Error:  2
Failed in %:  66.67%
```

As shown, it will tell you the failed path that the specific kustomization.yaml generated the error from. If you want to see the output from the kustomize command, you can use the `-v` flag. You can also combine `-v` with `-e` to only print out the paths that failed to build.

```bash
kustomize-validator ./_tests -ev

Validating Kustomization files ./_tests
[ERROR]: Error while executing kustomize in path: _tests/app1, validation failed: found 'PATCH_ME' in line 23 for resource apps/v1/Deployment/<none>/my-app
==> Stdout (_tests/app1):
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: my-app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - image: my-app:latest
        name: my-app
        ports:
        - containerPort: 8080
        resources:
          limits:
            cpu: 500m
            memory: PATCH_ME
          requests:
            cpu: 250m
            memory: 64Mi

==> Stderr (_tests/app1):

[ERROR]: Error while executing kustomize in path: _tests/app2, validation failed: found 'PATCH_ME' in line 23 for resource apps/v1/Deployment/<none>/my-app
==> Stdout (_tests/app2):
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: my-app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - image: my-app:latest
        name: my-app
        ports:
        - containerPort: 8080
        resources:
          limits:
            cpu: 500m
            memory: PATCH_ME
          requests:
            cpu: 250m
            memory: 64Mi

==> Stderr (_tests/app2):

Total:  3
Success:  1
Error:  2
Failed in %:  66.67%
```

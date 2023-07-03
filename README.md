# Kustomization.yaml valdidator
Simple script that will validate that all the _kustomization.yaml_ files are buildable. Build with:
`go build kustomization-validator`

## How to use it
Run the script from directory that you want to start to recursively traverse.

```sh
./kustomization-validator
```

## Example output
```sh
➜ ./kustomization-validator
OK: ../base/odf/resources/rook-ceph-operator-configmap
OK: ../base/advanced-cluster-security/init-bundle
OK: ../base/cluster-monitoring
OK: ../base/advanced-cluster-security/central
OK: ../base/chrony-configuration
OK: ../base/ldap-rbac
OK: ../base/odf/resources/namespace
FAILED: ../base/odf
Failed path: ../base/odf/resources/namespace.yaml
OK: ../base/managed-cluster/ch05-lkd01
OK: ../base/image-registry
.......
.......
====================
Total OK: 94
Total FAILED: 1
```
As shown, it will tell you the failed path that the specific kustomization.yaml generated the error from.


## Use in CI
This can be used to validate kustomization.yaml when pushing code to a git server, e.g. GitHub. It works well together with a linting tool, such as [yamllint](https://yamllint.readthedocs.io/en/stable/#).

To set it up with a runner, you need the following to be installed:

In order to use this workflow, these things are required:
* Access token with sufficient permissions to the repository
* Runner with the following installed:
    * [Helm](https://helm.sh/docs/intro/install/)
    * [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)
    * [PolicyGenerator plugin](https://github.com/stolostron/policy-generator-plugin#prerequisite)

All of these binaries need to be added to the runners $PATH
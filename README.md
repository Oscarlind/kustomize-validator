# Kustomization.yaml valdidator

A simple script that will validate that all of the Kustomization files are buildable. This script expects a base directory to start the recursive search from. It will then look for all the kustomization.yaml files and try to build them. If the build fails, it will print out the path to the failed kustomization.yaml file.

## Prerequisites

* Go 1.23 or later
* [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) v5.6.0 or later
* [Helm](https://helm.sh/docs/intro/install/) v3.16.4 or later

## Building the script

```bash
go build -o kustomization-validator main.go
```

## How to use it

Run the script from any location, and provide the base directory as an argument. The script will then recursively search for all kustomization.yaml or kustomization.yml files and try to build them.

```sh
./kustomization-validator ./base/path
```

Optional flags:

* `-v` or `--verbose` - Will print out the output from the kustomize command
* `-e` or `--error-only` - Will only print out the paths that failed to build

## Example output

```bash
Validating Kustomization files /base/path
[OK]: Successfully executed kustomize on /base/path/folder1
[ERROR]: Error while executing kustomize in path: /base/path/folder2, exit status 1
Total:  2
Success:  1
Error:  1
Failed in %:  50.0%
```

As shown, it will tell you the failed path that the specific kustomization.yaml generated the error from. If you want to see the output from the kustomize command, you can use the `-v` flag. You can also combine `-v` with `-e` to only print out the paths that failed to build.

```bash
Validating Kustomization files /base/path
ERROR: Error while executing kustomize in path: /base/path/folder2, exit status 1
==> Stdout (/base/path/folder2):

==> Stderr (/base/path/folder2:
Error: add operation does not apply: doc is missing path: "spec/filters/-": missing value

Total:  2
Success:  1
Error:  1
Failed in %:  50.0%
```

## Use in CI

This can be used to validate kustomization.yaml when pushing code to a git server, e.g. GitHub. It works well together with a linting tool, such as [yamllint](https://yamllint.readthedocs.io/en/stable/#).

To set it up with a runner, you need the following to be installed:

In order to use this workflow, these things are required:

* Access token with sufficient permissions to the repository
* Runner with the prerequisites installed

All of these binaries need to be added to the runners $PATH environment variable.

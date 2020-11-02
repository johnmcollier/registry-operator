# Devfile Registry Operator

Devfile Registry operator repository that contains the operator for the DevfileRegistry Custom Resource. 

## Issue Tracking

Issue tracking repo: https://github.com/devfile/api with label area/registry

## Running the controller in a cluster

The controller can be deployed to a cluster provided you are logged in with cluster-admin credentials:

```bash
export IMG=johncollier/registry-operator:v0.0.1
export TOOL=oc # Use 'export TOOL=kubectl' for kubernetes
make install && make deploy
```

## Development

The repository contains a Makefile; building and deploying can be configured via the environment variables

|variable|purpose|default value|
|---|---|---|
| `IMG` | Image used for controller | `johncollier/registry-operator:v0.0.1` |
| `TOOL` | CLI tool for interfacing with the cluster: `kubectl` or `oc`; if `oc` is used, deployment is tailored to OpenShift, otherwise Kubernetes | `oc` |

Some of the rules supported by the makefile:

|rule|purpose|
|---|---|
| docker-build | build registry operator docker image |
| docker-push | push registry operator docker image |
| deploy | deploy operator to cluster |
| install | create the devfile registry operator CRDs on the cluster |
| manifests | Generate manifests e.g. CRD, RBAC etc. |
| generate | Generate the API type definitions. Must be run after modifying the DevfileRegistry type. |

To see all rules supported by the makefile, run `make help`

### Test run controller
1. Take a look samples workspace configuration in `./samples` folder.
2. Apply any of them by executing `kubectl apply -f ./samples/workspace_java_mysql.yaml -n <namespace>`
3. As soon as workspace is started you're able to get IDE url by executing `kubectl get devworkspace -n <namespace>`

### Run operator locally
It's possible to run an instance of the operator locally while communicating with a cluster. 

```bash
export NAMESPACE=devfileregistry-operator
export TOOL=oc # Use 'export TOOL=kubectl' for kubernetes
make run ENABLE_WEBHOOKS=false
```

When running locally, only a single namespace is watched; as a result, all workspaces have to be deployed to `${NAMESPACE}`
![TriggerMesh Logo](docs/assets/images/triggermesh-logo.png)

![CodeQL](https://github.com/triggermesh/triggermesh-core/actions/workflows/codeql.yaml/badge.svg?branch=main)
![Static](https://github.com/triggermesh/triggermesh-core/actions/workflows/static.yaml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/triggermesh/triggermesh-core)](https://goreportcard.com/report/github.com/triggermesh/triggermesh-core)
[![Release](https://img.shields.io/github/v/release/triggermesh/triggermesh-core?label=release)](https://github.com/triggermesh/triggermesh-core/releases)
[![Slack](https://img.shields.io/badge/Slack-Join%20chat-4a154b?style=flat&logo=slack)](https://join.slack.com/t/triggermesh-community/shared_invite/zt-1kngevosm-MY7kqn9h6bT08hWh8PeltA)

The TriggerMesh Core components conform the basis for creating event driven applications declaratively at Kubernetes.

## Installation

To install TriggerMesh Core at a Kubernetes cluster apply manifests for both CRDs and Controller:

```console
# Install TriggerMesh Core CRDs
kubectl apply -f https://github.com/triggermesh/triggermesh-core/releases/latest/download/triggermesh-core-crds.yaml

# Install TriggerMesh Core Controller
kubectl apply -f https://github.com/triggermesh/triggermesh-core/releases/latest/download/triggermesh-core.yaml
```

Refer to [releases](https://github.com/triggermesh/triggermesh-core/releases) for further information.

### Development Version

Development version can be installed using [ko](https://github.com/ko-build/ko)

```console
ko apply -f ./config
```

### Namespaced installation

TriggerMesh Core controller can be configured to work with a single namespace set at the `WORKING_NAMESPACE` environment variable, which can be added editing the deployment manifest.

```yaml
        - name: WORKING_NAMESPACE
          value: my-namespace
```

When working with a single namespace, all `ClusterRoleBindings` should also be modified adding the namespace to limit the scope of the granted permissions.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: triggermesh-core-controller
  namespace: working
...
```

## Concepts

TriggerMesh core contains Kubernetes objects for Brokers and Triggers:

- [RedisBroker](docs/redis-broker.md)
- [MemoryBroker](docs/memory-broker.md)
- [Trigger](docs/trigger.md)

The brokers are used to ingest events and route them to targets. To ingest events, they must conform to the [CloudEvents specification][ce-spec] using the HTTP binding, and must use the HTTP address exposed by the Broker.

Events consumption is done asynchronously by configuring Triggers that reference a Broker object. A Trigger must also include information about the consumer address, either a Kubernetes object or an HTTP address, and optionally can include an event filter.

## Usage

- [Getting Started (Redis Broker)](docs/getting-started-redis.md).
- [Getting Started (Memory Broker)](docs/getting-started-memory.md).
- [Broker Observability](docs/observable-broker.md).

## Contributing

Please refer to our [guidelines for contributors](CONTRIBUTING.md).

## Commercial Support

TriggerMesh Inc. offers commercial support for the TriggerMesh platform. Email us at <info@triggermesh.com> to get more
details.

## License

This software is licensed under the [Apache License, Version 2.0][asl2].

Additionally, the End User License Agreement included in the [`EULA.pdf`](EULA.pdf) file applies to compiled
executables and container images released by TriggerMesh Inc.

[asl2]: https://www.apache.org/licenses/LICENSE-2.0
[ce-spec]: https://github.com/cloudevents/spec

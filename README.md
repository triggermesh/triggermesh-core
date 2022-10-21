![TriggerMesh Logo](docs/assets/triggermesh-logo.png)

![CodeQL](https://github.com/triggermesh/triggermesh-core/actions/workflows/codeql.yaml/badge.svg?branch=main)
![Static](https://github.com/triggermesh/triggermesh-core/actions/workflows/static.yaml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/triggermesh/triggermesh-core)](https://goreportcard.com/report/github.com/triggermesh/triggermesh-core)

The TriggerMesh Core components conform the basis for creating event driven applications declaratively at Kubernetes.

## Getting Started

TriggerMesh Core includes 2 components:

* RedisBroker, which uses a backing Redis instance to store events and routes them via Triggers.
* Trigger, which subscribes to events and push them to your targets.

Events must conform to [CloudEvents spec](https://github.com/cloudevents/spec) using the [HTTP binding](https://github.com/cloudevents/spec/blob/main/cloudevents/bindings/http-protocol-binding.md).

Create a RedisBroker named `demo`.

```console
kubectl apply -f https://raw.githubusercontent.com/triggermesh/triggermesh-core/main/docs/getting-started/broker.yaml
```

Wait until the RedisBroker is ready. It will inform in its status of the URL where events can be ingested.

```console
kubectl get redisbroker demo

NAME   URL                                                        AGE   READY   REASON
demo   http://demo-redisbroker-broker.default.svc.cluster.local   10s   True
```

To be able to use the broker we will create a Pod that allow us to send events inside the Kubernetes cluster.

```console
kubectl apply -f https://raw.githubusercontent.com/triggermesh/triggermesh-core/main/docs/getting-started/curl.yaml
```

It is possible now to send events to the broker address by issuing curl commands. The response for ingested events must be an `HTTP 200`.

```console
kubectl exec -ti curl -- curl -v http://demo-redisbroker-broker.default.svc.cluster.local:8080/ \
    -X POST \
    -H "Ce-Id: 1234-abcd" \
    -H "Ce-Specversion: 1.0" \
    -H "Ce-Type: demo.type1" \
    -H "Ce-Source: curl" \
    -H "Content-Type: application/json" \
    -d '{"hello":"world"}'
```

Sockeye is a popular CloudEvents consumer that exposes a web interface with the list of events received while the page is open. We will be creating 2 instances of sockeye, one as the target for consumed events and another one for the Dead Letter Sink.
A Dead Letter Sink, abbreviated DLS is a destination that consumes events that a subscription was not able to deliver.

```console
# Target service
kubectl apply -f https://raw.githubusercontent.com/triggermesh/triggermesh-core/main/docs/getting-started/sockeye-target.yaml

# DLS service
kubectl apply -f https://raw.githubusercontent.com/triggermesh/triggermesh-core/main/docs/getting-started/sockeye-deadlettersink.yaml
```

The Trigger object configures the broker to consume events and send them to a target. The Trigger object can include filters that select which events should be forwarded to the target, and delivery options to configure retries and fallback targets when the event cannot be delivered.

```console
kubectl apply -f https://raw.githubusercontent.com/triggermesh/triggermesh-core/main/docs/getting-started/trigger.yaml
```

The Trigger created above filters by CloudEvents with `type: demo.type1`, if delivery fails it will issue 3 retries and then forward the CloudEvent to the `sokceye-deadlettersink` service.

Using the `curl` Pod again we can send this CloudEvent to the broker, that will pass the filtering and forward the event to the `sockeye-target` service.

```console
kubectl exec -ti curl -- curl -v http://demo-redisbroker-broker.default.svc.cluster.local:8080/ \
    -X POST \
    -H "Ce-Id: 1234-abcd" \
    -H "Ce-Specversion: 1.0" \
    -H "Ce-Type: demo.type1" \
    -H "Ce-Source: curl" \
    -H "Content-Type: application/json" \
    -d '{"hello":"world"}'
```

## Installation

Devevlopment version might be unstable.

```console
ko apply -f ./config
```

## Usage

### Brokers

TODO

### Triggers

TODO

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

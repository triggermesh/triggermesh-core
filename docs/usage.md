# Usage

This is a TODO doc page containing examples.

## RedisBroker

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: RedisBroker
metadata:
  name: mybroker
```

## RedisBroker using external Redis

```console
kubectl create secret generic redis-creds \
  --from-literal=username=user1 \
  --from-literal=password=password1
```

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: RedisBroker
metadata:
  name: mybroker-external
spec:
  redis:
    connection:
      url: testme.com:9876
      username:
        secretKeyRef:
          name: redis-creds
          key: username
      password:
        secretKeyRef:
          name: redis-creds
          key: password
    stream: mybroker-stream
    streamMaxLen: 1000
  broker:
    # port defaults to 80
    port: 1080
```


## Triggers

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: Trigger
metadata:
  name: my-trigger
spec:
  broker:
    kind: RedisBroker
    group: eventing.triggermesh.io
    name: my-broker
  filters:
  - any:
    - exact:
        type: my.type1
    - exact:
        type: my.type2
  target:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: eventdisplay
  delivery:
    retry: 3
    backoffPolicy: linear
    backoffDelay: PT5S
    deadLetterSink:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: eventdisplay-dls
```

- **spec.broker** needs to be informed the broker instance to subscribe to. It can use either `group, kind, name` or  `apiVersion, kind, name`.
- **spec.filters** (optional) contains a dialect (or a hierarchy of them) that will be evaulated against all CloudEvents flowing through the broker, trying to send to the target only those that pass the filter.
- **spec.target** needs to be informed either a reference to a Kubernetes object or a URI.
- **delivery.retry** (optional) minimum number of retries for delivering each message.
- **delivery.backoffPolicy** retries policy, can be linear or exponential.
- **delivery.backoffDelay** delay before retries using [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) duration format.
- **delivery.deadLetterSink** (optional) points to a Kubernetes object or URI to send not delivered messages.

When using Kubernetes objects for a target or DLS destination the `triggermesh-core` controller must have permissions to `get, list, watch` those objects, which is usually done by creating a ClusterRole for that contains the `duck.knative.dev/addressable: "true"` label.

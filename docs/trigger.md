# Trigger

Triggers objects build subscriptions from Brokers to event consumers. A Trigger contains 3 elements: broker reference, target and filter.

- The referenced broker will be reconfigured to address for the new subscription.
- The target must be set to either a Kubernetes object that can be resolved to an URL, or a URL. In either case the target must expose an HTTP endpoint to receive events.
- The filter declaratively selects which events should be forwarded to the target. In absence of filter all messages received at the Broker will be sent to the configured Target.

## Spec

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: Trigger
metadata:
  name: <trigger name>
spec:
  broker:
    apiVersion: <Kubernetes apiVersion for the Broker object. Can inform 'group' instead>
    group: <Kubernetes group for the Broker object. Can inform 'apiVersion' instead>
    kind: <Kubernetes kind for the Broker object>
    name: <name of the Broker object>
  target: <Destination where events will be sent. Either reference to an objet or URI>
    ref:
      apiVersion: <Kubernetes apiVersion for the consumer object. Can inform 'group' instead>
      group: <Kubernetes group for the consumer object. Can inform 'apiVersion' instead>
      kind: <Kubernetes kind for the consumer object>
      name: <name of the consumer object>
    uri: <URI to the event consumer HTTP endpoint>
  delivery: <Event delivery options>
    retry: <Number of tries to deliver an event before considering failed>
    backoffDelay: <Backoff duration factor between retries>
    backoffPolicy: <Backoff policy applied to the delay, can be linear, exponential or constant>
    deadLetterSink: <Destination where underlivered events will be sent>
      ref:
        apiVersion: <Kubernetes apiVersion for the DLS object. Can inform 'group' instead>
        group: <Kubernetes group for the DLS object. Can inform 'apiVersion' instead>
        kind: <Kubernetes kind for the DLS object>
        name: <name of the DLS object>
      uri: <URI to the event DLS HTTP endpoint>
  filters: <Filter specification. See 'Filtering Events' section in this doc>
  bounds: <Event offsets that this trigger should be retrieving>
    byId: <Offsets defined by the broker's backend event identifiers>
      start: <Starting offset>
      end: <Ending offset>
    byDate: <Offsets defined by timestamps formatted as RFC3339>
      start: <Starting offset>
      end: <Ending offset>
```

- `spec.broker` must be a running broker that will be configured with this Trigger's configuration.
- `spec.target` must refer to an endpoint that will receive events from the Broker. When the event consumer is a Kubernetes object it is prefered to use the `spec.target.ref` structure.
- `spec.delivery` contains the logic to apply when an event cannot be delivered from the Broker to a Target, performing a number of retries, and finally sending to a dead letter sink if none of them succeed. Duration format for `spec.
- `spec.filters` contains a set of filter expresions. See the [Filtering Events section](#filtering-events)
- `spec.bounds` contains optional start and end offsets for the event that the Trigger is intereseted in receiving. When using dates, [RFC3339 format](https://utcc.utoronto.ca/~cks/space/blog/unix/GNUDateAndRFC3339) should be used.

## Filtering Events

Events flowing through a Broker can be filtered before being sent to targets by using a range of expressions. TriggerMesh filter supports the [CloudEvents Subscriptions API filters](https://github.com/cloudevents/spec/blob/main/subscriptions/spec.md#324-filters), but will extend it with custom _dialects_ in the future.

There are 2 categories of filter dialects, those that perform actual filtering and those that group other dialects adding boolean logic to the inner dialects:

These filter dialects are meant to group inner dialects:

- `all` contains an array of dialects, all of them must resolve to true for positive matching.
- `any` contains an array of dialects, at least one of them must resolve to true for positive matching.
- `not` contains an array of dialects, all of them must resolve to false for positive matching.

These filter dialects contain matching logic applied to the [CloudEvent attributes](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md#context-attributes):

- `exact` contains a key/value pair. For positive matching an attribute by that key/value must exist at the event.
- `prefix` contains key/value pair. For positive matching an attribute by the exact key that contains a value that is prefixed by the one at the filter must exist at the event.
- `suffix` contains a key/value pair. For positive matching an attribute by the exact key that contains a value that is suffixed by the one at the filter must exist at the event.

### Examples

Filter for events whose `type` attribute is set to `io.triggermesh.demo`

```yaml
filters:
- exact:
    type: my.demo.type
```

Filter for events whose `type` attribute is set to `io.triggermesh.demo` and `category` is set to `test`

```yaml
filters:
- all:
  - exact:
      type: io.triggermesh.demo
  - exact:
      category: test
```

Filter for events whose `type` attribute starts with `io.triggermesh.` or `type` is set to `io.tm.demo`

```yaml
filters:
- any:
  - prefix:
      type: io.triggermesh.
  - exact:
      type: io.tm.demo
```

Filter for events whose `type` attribute does not ends with `.avoid.me` or `.avoid.me.too`.

```yaml
filters:
- not:
  - any:
    - suffix:
        type: .avoid.me
    - suffix:
        type: .avoid.me.too
```

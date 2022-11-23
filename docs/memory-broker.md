# Memory Broker

The `MemoryBroker` is a very simple and effective Broker that do not persist events. For stronger delivery guarantees see [RedisBroker](redis-broker.md).

## Spec

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: MemoryBroker
metadata:
  name: <broker instance name>
spec:
  memory:
    bufferSize: <maximum number of events>
  broker:
    port: <HTTP port for ingesting events>
    observability:
      valueFromConfigMap: <kubernetes ConfigMap that contains observability configuration>
```

The only `MemoryBroker` specific parameter is `spec.memory.bufferSize` which indicates the availible size of the internal queue that the broker manages. When the maximum number of items is reached, new ingest requests will block and might eventually time out. This parameter is optional and defaults to 10000.

The `spec.broker` section contains generic Borker parameters:

- `spec.broker.port` that the Broker service will be listening at. Optional, defaults to port 80.
- `spec.broker.observability` can be set to the name of a ConfigMap at the same namespace that contains [observability settings](observability.md). This parameter is optional.

## Example

- See [MemoryBroker example](https://github.com/triggermesh/triggermesh-core/blob/main/docs/assets/manifests/getting-started-memory/broker.yaml)
- See [MemoryBroker getting started guide](getting-started-memory.md)

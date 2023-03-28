# Redis Broker

The `RedisBroker` persist every ingested message at a Stream before returning an ACK to the event sender, making it more reliable than the [MemoryBroker](memory-broker.md).

It can be configured with any Redis instance version 6 and up, by providing connection parameters. If a Redis connection is not informed a Redis Deployment will be created by the TriggerMesh Core controller.

## Spec

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: RedisBroker
metadata:
  name: <broker instance name>
spec:
  redis:
    connection: <Provides a connection to an external Redis instance. Optional>
        url: <redis URL. Required if clusterURLs not informed>
        clusterURLs:
        - <an entry for each redis URL in the cluster. Required if url not informed>
        username: <redis username, referenced using a Kubernetes secret>
          secretKeyRef:
            name: <Kubernetes secret name>
            key: <Kubernetes secret key>
        password: <redis password, referenced using a Kubernetes secret>
          secretKeyRef:
            name: <Kubernetes secret name>
            key: <Kubernetes secret key>
        caCertificate: <CA certificate used to connect to redis. Optional>
          secretKeyRef:
            name: <Kubernetes secret name>
            key: <Kubernetes secret key>
        tlsEnabled: <boolean that indicates if the Redis server is TLS protected. Optional, defaults to false>
        tlsSkipVerify: <boolean that skips verifying TLS certificates. Optional, defaults to false>
    stream: <Redis stream name. Optional, defaults to a combination of namespace and broker name>
    streamMaxLen: <maximum number of items the Redis stream can host. Optional, defaults to unlimited>
  broker:
    port: <HTTP port for ingesting events>
    observability:
      valueFromConfigMap: <kubernetes ConfigMap that contains observability configuration>
```

The only `RedisBroker` specific parameters are:

- `spec.redis.connection`. When not used the broker will spin up a managed Redis Deployment. However for production scenarios that require HA and hardened security it is recommended to provide the connection to a user managed Redis instance.
- `spec.stream` is the Redis stream name to be used by the broker. If it doesn't exists the Broker will create it.
- `spec.streamMaxLen` is the maximum number of elements that the stream will contain.

The `spec.broker` section contains generic Borker parameters:

- `spec.broker.port` that the Broker service will be listening at. Optional, defaults to port 80.
- `spec.broker.observability` can be set to the name of a ConfigMap at the same namespace that contains [observability settings](observability.md). This parameter is optional.

## Example

- See [RedisBroker example](https://github.com/triggermesh/triggermesh-core/blob/main/docs/assets/manifests/getting-started-redis/broker.yaml)
- See [RedisBroker getting started guide](getting-started-redis.md)

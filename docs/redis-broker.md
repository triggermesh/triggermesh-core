# Redis Broker

## Spec

```yaml
apiVersion: eventing.triggermesh.io/v1alpha1
kind: RedisBroker
metadata:
  name: <broker instance name>
spec:
  redis:
    connection: <Provides a connection to an external Redis instance. Optional>
        url: <redis URL. Required>
        username: <redis username, referenced using a Kubernetes secret>
          secretKeyRef:
            name: <Kubernetes secret name>
            key: <Kubernetes secret key>
        password: <redis password, referenced using a Kubernetes secret>
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

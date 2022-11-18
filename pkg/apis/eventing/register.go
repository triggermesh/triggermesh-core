package eventing

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	GroupName = "eventing.triggermesh.io"
)

var (
	// BrokersResource represents a TriggerMesh Redis Broker
	RedisBrokersResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "redisbrokers",
	}

	// BrokersResource represents a TriggerMesh Memory Broker
	MemoryBrokersResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "memorybrokers",
	}
)

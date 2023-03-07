package resources

const (
	TestBrokerImage = "image.test:v.test"
	TestNamespace   = "test-namespace"
	TestName        = "test-name"
)

var (
	TestTrue           = true
	TestReplicas int32 = 1
)

type BrokerHelper struct {
	Suffix string
	Kind   string
}

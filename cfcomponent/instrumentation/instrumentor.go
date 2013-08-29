package instrumentation

type Instrumentable interface {
	Emit() Context
	Metrics() []Metric
}

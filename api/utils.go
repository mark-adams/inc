package api

// MetricCollector reports application metrics
type MetricCollector interface {
	Increment(bucket string)
}

// NullMetricsCollector is a stub metric reporter that doesn't do anything with the metrics
type NullMetricsCollector struct{}

// Increment increases a metric
func (n *NullMetricsCollector) Increment(bucket string) {
	// Don't do anything
}

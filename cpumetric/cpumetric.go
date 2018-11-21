package cpumetric

import "github.com/cloudfoundry/sonde-go/events"

type MetricType int

const (
	Empty MetricType = iota
	Usage
	Entitlement
)

func newMetricType(name string) MetricType {
	switch name {
	case "absolute_usage":
		return Usage
	case "absolute_entitlement":
		return Entitlement
	default:
		return Empty
	}
}

type CpuMetric struct {
	Type      MetricType
	Name      string
	Value     float64
	Timestamp int64
}

func FromEnvelope(envelope *events.Envelope) CpuMetric {
	if envelope.ValueMetric == nil {
		return CpuMetric{}
	}

	if envelope.ValueMetric.Value == nil {
		return CpuMetric{}
	}

	if envelope.Timestamp == nil {
		return CpuMetric{}
	}

	if envelope.ValueMetric.Name == nil {
		return CpuMetric{}
	}

	t := newMetricType(*envelope.ValueMetric.Name)
	if t == Empty {
		return CpuMetric{}
	}

	return CpuMetric{Type: t, Value: *envelope.ValueMetric.Value, Timestamp: *envelope.Timestamp}
}

func Aggregate(metrics <-chan CpuMetric, outputs chan<- float64) {
	var (
		values    = map[string]float64{}
		timestamp int64
	)

	for metric := range metrics {
		values[metric.Name] = metric.Value
		if metric.Timestamp == timestamp {
			outputs <- values["absolute_usage"] / values["absolute_entitlement"]
		} else {
			values = map[string]float64{metric.Name: metric.Value}
			timestamp = metric.Timestamp
		}
	}
}

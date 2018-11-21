package cpumetric

import "github.com/cloudfoundry/sonde-go/events"

type MetricType int

const (
	Empty MetricType = iota
	Usage
	Entitlement
	Age
)

func newMetricType(name string) MetricType {
	switch name {
	case "absolute_usage":
		return Usage
	case "absolute_entitlement":
		return Entitlement
	case "container_age":
		return Age
	default:
		return Empty
	}
}

type CPUMetric struct {
	Type      MetricType
	Value     float64
	Timestamp int64
}

func New(t MetricType, value float64, timestamp int64) CPUMetric {
	return CPUMetric{Type: t, Value: value, Timestamp: timestamp}
}

func FromEnvelope(envelope *events.Envelope) CPUMetric {
	if envelope.ValueMetric == nil {
		return CPUMetric{}
	}

	if envelope.ValueMetric.Value == nil {
		return CPUMetric{}
	}

	if envelope.Timestamp == nil {
		return CPUMetric{}
	}

	if envelope.ValueMetric.Name == nil {
		return CPUMetric{}
	}

	t := newMetricType(*envelope.ValueMetric.Name)
	if t == Empty {
		return CPUMetric{}
	}

	return CPUMetric{Type: t, Value: *envelope.ValueMetric.Value, Timestamp: *envelope.Timestamp}
}

func Aggregate(metrics <-chan CPUMetric, outputs chan<- float64) {
	var (
		values    = make(map[MetricType]float64)
		timestamp int64
	)

	for metric := range metrics {
		values[metric.Type] = metric.Value
		if metric.Timestamp == timestamp {
			outputs <- values[Usage] / values[Entitlement]
		} else {
			values = make(map[MetricType]float64)
			values[metric.Type] = metric.Value
			timestamp = metric.Timestamp
		}
	}
}

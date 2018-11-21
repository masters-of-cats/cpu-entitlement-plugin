package cpumetric_test

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/masters-of-cats/cpu-entitlement-plugin/cpumetric"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CalculateCpu", func() {
	var (
		inputs  chan cpumetric.CpuMetric
		outputs chan float64
	)

	BeforeEach(func() {
		inputs = make(chan cpumetric.CpuMetric)
		outputs = make(chan float64, 1)
		go cpumetric.Aggregate(inputs, outputs)
	})

	AfterEach(func() {
		close(inputs)
	})

	It("returns nothings if not enough data", func() {
		inputs <- cpumetric.CpuMetric{Name: "absolute_usage", Value: 1, Timestamp: 1}
		Consistently(outputs).ShouldNot(Receive())
	})

	It("returns a cpu percentage", func() {
		inputs <- cpumetric.CpuMetric{Name: "container_age", Value: 2323, Timestamp: 1}
		inputs <- cpumetric.CpuMetric{Name: "absolute_usage", Value: 1, Timestamp: 1}
		inputs <- cpumetric.CpuMetric{Name: "absolute_entitlement", Value: 2, Timestamp: 1}
		var result float64
		Eventually(outputs).Should(Receive(&result))
		Expect(result).To(Equal(0.5))
	})

	It("handles dropped messages", func() {
		inputs <- cpumetric.CpuMetric{Name: "absolute_usage", Value: 2, Timestamp: 1}
		inputs <- cpumetric.CpuMetric{Name: "absolute_entitlement", Value: 3, Timestamp: 2}
		inputs <- cpumetric.CpuMetric{Name: "absolute_usage", Value: 3, Timestamp: 2}
		var result float64
		Eventually(outputs).Should(Receive(&result))
		Expect(result).To(Equal(1.0))
	})

	FDescribe("FromEnvelope", func() {
		var envelope *events.Envelope

		BeforeEach(func() {
			envelope = newEnvelope(1, newValueMetric("absolute_usage", 1))
		})

		It("generates a new metric", func() {
			metric := cpumetric.FromEnvelope(envelope)
			expected := cpumetric.CpuMetric{Type: cpumetric.Usage, Value: 1, Timestamp: 1}
			Expect(expected).To(Equal(metric))
		})

		// When("supplied an uninsteresting envelope", func() {
		//
		// })
		//
		// It("returns notok for incorrect envelope types", func() {})
	})
})

func newValueMetric(name string, value float64) *events.ValueMetric {
	return &events.ValueMetric{Name: stringptr("absolute_usage"), Value: float64ptr(1)}
}

func newEnvelope(timestamp int64, valueMetric *events.ValueMetric) *events.Envelope {
	return &events.Envelope{Timestamp: int64ptr(timestamp), ValueMetric: valueMetric}
}

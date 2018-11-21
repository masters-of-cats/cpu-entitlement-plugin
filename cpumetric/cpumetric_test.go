package cpumetric_test

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/masters-of-cats/cpu-entitlement-plugin/cpumetric"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CalculateCpu", func() {
	var (
		inputs  chan cpumetric.CPUMetric
		outputs chan float64
	)

	BeforeEach(func() {
		inputs = make(chan cpumetric.CPUMetric)
		outputs = make(chan float64, 1)
		go cpumetric.Aggregate(inputs, outputs)
	})

	AfterEach(func() {
		close(inputs)
	})

	It("returns nothings if not enough data", func() {
		inputs <- cpumetric.New(cpumetric.Usage, 1, 1)
		Consistently(outputs).ShouldNot(Receive())
	})

	It("returns a cpu percentage", func() {
		inputs <- cpumetric.New(cpumetric.Age, 2323, 1)
		inputs <- cpumetric.New(cpumetric.Usage, 1, 1)
		inputs <- cpumetric.New(cpumetric.Entitlement, 2, 1)
		var result float64
		Eventually(outputs).Should(Receive(&result))
		Expect(result).To(Equal(0.5))
	})

	It("handles dropped messages", func() {
		inputs <- cpumetric.New(cpumetric.Usage, 2, 1)
		inputs <- cpumetric.New(cpumetric.Entitlement, 3, 2)
		inputs <- cpumetric.New(cpumetric.Usage, 3, 2)
		var result float64
		Eventually(outputs).Should(Receive(&result))
		Expect(result).To(Equal(1.0))
	})

	FDescribe("CPUMetric", func() {
		Describe("FromEnvelope", func() {
			var (
				envelope *events.Envelope
				metric   cpumetric.CPUMetric
			)

			BeforeEach(func() {
				envelope = newEnvelope(1, newValueMetric("absolute_usage", 1))
			})

			JustBeforeEach(func() {
				metric = cpumetric.FromEnvelope(envelope)
			})

			It("generates a new metric", func() {
				expected := cpumetric.CPUMetric{Type: cpumetric.Usage, Value: 1, Timestamp: 1}
				Expect(expected).To(Equal(metric))
			})

			When("supplied an uninsteresting envelope", func() {
				BeforeEach(func() {
					envelope = newEnvelope(1, newValueMetric("boring", 1))
				})

				It("returns an empty metric", func() {
					expected := cpumetric.CPUMetric{Type: cpumetric.Empty}
					Expect(expected).To(Equal(metric))
				})
			})

			When("ValueMetric is nil", func() {
				BeforeEach(func() {
					envelope.ValueMetric = nil
				})

				It("returns an empty metric", func() {
					expected := cpumetric.CPUMetric{Type: cpumetric.Empty}
					Expect(expected).To(Equal(metric))
				})
			})

			When("ValueMetric.Name is nil", func() {
				BeforeEach(func() {
					envelope.ValueMetric.Name = nil
				})

				It("returns an empty metric", func() {
					expected := cpumetric.CPUMetric{Type: cpumetric.Empty}
					Expect(expected).To(Equal(metric))
				})
			})

			When("ValueMetric.Value is nil", func() {
				BeforeEach(func() {
					envelope.ValueMetric.Value = nil
				})

				It("returns an empty metric", func() {
					expected := cpumetric.CPUMetric{Type: cpumetric.Empty}
					Expect(expected).To(Equal(metric))
				})
			})

			When("ValueMetric.Name is container_age", func() {
				BeforeEach(func() {
					envelope.ValueMetric.Name = stringptr("container_age")
				})

				It("generates the correct metric type", func() {
					expected := cpumetric.CPUMetric{Type: cpumetric.Age, Value: 1, Timestamp: 1}
					Expect(expected).To(Equal(metric))
				})
			})

			When("ValueMetric.Name is absolute_entitlement", func() {
				BeforeEach(func() {
					envelope.ValueMetric.Name = stringptr("absolute_entitlement")
				})

				It("generates the correct metric type", func() {
					expected := cpumetric.CPUMetric{Type: cpumetric.Entitlement, Value: 1, Timestamp: 1}
					Expect(expected).To(Equal(metric))
				})
			})
		})
	})
})

func newValueMetric(name string, value float64) *events.ValueMetric {
	return &events.ValueMetric{Name: stringptr(name), Value: float64ptr(1)}
}

func newEnvelope(timestamp int64, valueMetric *events.ValueMetric) *events.Envelope {
	return &events.Envelope{Timestamp: int64ptr(timestamp), ValueMetric: valueMetric}
}

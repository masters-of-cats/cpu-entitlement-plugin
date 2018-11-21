package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CalculateCpu", func() {
	var (
		inputs  chan CpuMetric
		outputs chan float64
	)

	BeforeEach(func() {
		inputs = make(chan CpuMetric)
		outputs = Calculate(inputs)
	})

	It("returns nothings if not enough data", func() {
		inputs <- CpuMetric{Name: "absolute_usage", Value: 1, Timestamp: 1}
		Consistently(outputs).ShouldNot(Receive())
	})

	It("returns a cpu percentage", func() {
		inputs <- CpuMetric{Name: "absolute_usage", Value: 1, Timestamp: 1}
		inputs <- CpuMetric{Name: "absolute_entitlement", Value: 2, Timestamp: 1}
		inputs <- CpuMetric{Name: "container_age", Value: 2323, Timestamp: 1}
		var result float64
		Eventually(outputs).Should(Receive(&result))
		Expect(result).To(Equal(0.5))
	})

	It("handles dropped messages", func() {
		inputs <- CpuMetric{Name: "absolute_usage", Value: 2, Timestamp: 1}
		inputs <- CpuMetric{Name: "absolute_entitlement", Value: 3, Timestamp: 2}
		inputs <- CpuMetric{Name: "absolute_usage", Value: 3, Timestamp: 2}
		var result float64
		Eventually(outputs).Should(Receive(&result))
		Expect(result).To(Equal(1.0))
	})
})

package cpumetric_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCpuEntitlementPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CpuEntitlementPlugin Suite")
}

func int64ptr(i int64) *int64 {
	return &i
}

func stringptr(s string) *string {
	return &s
}

func float64ptr(f float64) *float64 {
	return &f
}

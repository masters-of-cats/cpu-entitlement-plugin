package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCpuEntitlementPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CpuEntitlementPlugin Suite")
}

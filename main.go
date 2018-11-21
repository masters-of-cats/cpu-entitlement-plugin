package main

import (
	"code.cloudfoundry.org/cli/plugin"
	cpuplugin "github.com/masters-of-cats/cpu-entitlement-plugin/plugin"
)

func main() {
	plugin.Start(cpuplugin.CPUEntitlementPlugin{})
}

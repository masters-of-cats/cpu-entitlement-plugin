package main

import (
	"code.cloudfoundry.org/cli/plugin"
	cpuplugin "github.com/masters-of-cats/cpu-entitlement-plugin/plugin"
)

func main() {
	appName := "dora"
	plugin.Start(cpuplugin.New(appName))
}

package main

import (
	"os"

	"code.cloudfoundry.org/cli/plugin"
	cpuplugin "github.com/masters-of-cats/cpu-entitlement-plugin/plugin"
)

func main() {
	appName := os.Args[1]
	plugin.Start(cpuplugin.New(appName))
}

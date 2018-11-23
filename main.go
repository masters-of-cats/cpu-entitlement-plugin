package main

import (
	"os"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cpu-entitlement-plugin/logstreamer"
	cpuplugin "github.com/cloudfoundry/cpu-entitlement-plugin/plugin"
)

func main() {
	logStreamer := logstreamer.New()
	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	terminalUI := terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	cpuPlugin := cpuplugin.New(logStreamer, terminalUI)

	plugin.Start(cpuPlugin)
}

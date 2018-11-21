package main

import (
	"crypto/tls"
	"os"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/masters-of-cats/cpu-entitlement-plugin/cpumetric"
)

type CpuEntitlementPlugin struct{}

func main() {
	plugin.Start(new(CpuEntitlementPlugin))
}

func (p *CpuEntitlementPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	ui := terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	// TODO: some args checking
	appName := args[1]
	app, err := cliConnection.GetApp(appName)
	if err != nil {
		ui.Failed(err.Error())
	}

	dopplerEndpoint, err := cliConnection.DopplerEndpoint()
	if err != nil {
		ui.Failed(err.Error())
	}

	authToken, err := cliConnection.AccessToken()
	if err != nil {
		ui.Failed(err.Error())
	}

	dopplerConnection := consumer.New(dopplerEndpoint, &tls.Config{InsecureSkipVerify: true}, nil)
	envelopes, errors := dopplerConnection.FilteredFirehose("CpuEntitlementPlugin", authToken, consumer.Metrics)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for err := range errors {
			ui.Warn(err.Error())
			return
		}
	}()

	defer dopplerConnection.Close()

	ui.Say("Hit Ctrl+c to exit")

	inputs := make(chan cpumetric.CpuMetric)
	defer close(inputs)
	go func() {
		for envelope := range envelopes {
			metric := cpumetric.FromEnvelope(envelope)
			if metric.Type == cpumetric.Empty {
				continue
			}

			inputs <- metric
		}
	}()

	outputs := make(chan float64, 1)
	go cpumetric.Aggregate(inputs, outputs)
	for metric := range outputs {
		ui.Say("%v%%", metric*100)
	}
	<-done
}

func isValueMetric(envelope *events.Envelope) bool {
	return *envelope.EventType == events.Envelope_ValueMetric
}

func isFromApp(envelope *events.Envelope, app plugin_models.GetAppModel) bool {
	return envelope.Tags["source_id"] == app.Guid
}

func (p *CpuEntitlementPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CpuEntitlementPlugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 2,
		},
		Commands: []plugin.Command{
			{
				Name:     "cpu-entitlement",
				Alias:    "cpu",
				HelpText: "See cpu entitlement per app",
				UsageDetails: plugin.Usage{
					Usage: "cf cpu-entitlement APP_NAME",
				},
			},
		},
	}
}

package plugin

import (
	"crypto/tls"
	"os"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/masters-of-cats/cpu-entitlement-plugin/cpumetric"
	"github.com/masters-of-cats/cpu-entitlement-plugin/envelopefilter"
)

type CPUEntitlementPlugin struct{}

func (p CPUEntitlementPlugin) Run(cliConnection plugin.CliConnection, args []string) {
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

	appEnvelopes := make(chan *events.Envelope)
	defer close(appEnvelopes)

	go envelopefilter.Filter(
		envelopes,
		appEnvelopes,
		envelopefilter.ValueMetric,
		envelopefilter.OfApp(app),
	)

	ui.Say("Hit Ctrl+c to exit")

	metrics := make(chan cpumetric.CPUMetric)
	defer close(metrics)

	go func() {
		for envelope := range envelopes {
			metric := cpumetric.FromEnvelope(envelope)
			if metric.Type == cpumetric.Empty {
				continue
			}

			metrics <- metric
		}
	}()

	output := make(chan float64, 1)
	go cpumetric.Aggregate(metrics, output)
	for metric := range output {
		ui.Say("%v%%", metric*100)
	}
	<-done
}

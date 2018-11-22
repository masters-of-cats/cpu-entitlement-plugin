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

type CPUEntitlementPlugin struct {
	name string
	ui   terminal.UI
}

func New(name string) *CPUEntitlementPlugin {
	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")

	return &CPUEntitlementPlugin{
		name: name,
		ui:   terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger),
	}
}

func (p *CPUEntitlementPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	app, err := cliConnection.GetApp(p.name)
	if err != nil {
		p.ui.Failed(err.Error())
	}

	dopplerEndpoint, err := cliConnection.DopplerEndpoint()
	if err != nil {
		p.ui.Failed(err.Error())
	}

	authToken, err := cliConnection.AccessToken()
	if err != nil {
		p.ui.Failed(err.Error())
	}

	connection := consumer.New(dopplerEndpoint, &tls.Config{InsecureSkipVerify: true}, nil)
	envelopes, errors := connection.FilteredFirehose("CPUEntitlementPlugin2", authToken, consumer.Metrics)
	defer connection.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for err := range errors {
			p.ui.Warn(err.Error())
			return
		}
	}()

	appEnvelopes := make(chan *events.Envelope)

	go func() {
		defer close(appEnvelopes)
		envelopefilter.Filter(
			envelopes,
			appEnvelopes,
			envelopefilter.ValueMetric,
			envelopefilter.OfApp(app),
		)
	}()

	p.ui.Say("Hit Ctrl+c to exit")

	metrics := make(chan cpumetric.CPUMetric)
	defer close(metrics)

	go func() {
		for envelope := range envelopes {
			metrics <- cpumetric.FromEnvelope(envelope)
		}
	}()

	output := make(chan float64, 1)
	go func() {
		defer close(output)
		cpumetric.Aggregate(metrics, output)
	}()

	for metric := range output {
		p.ui.Say("%.2f%%", metric*100)
	}
	<-done
}

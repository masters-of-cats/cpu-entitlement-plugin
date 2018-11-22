package plugin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
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

func (p *CPUEntitlementPlugin) Run(cli plugin.CliConnection, args []string) {
	app, err := cli.GetApp(p.name)
	if err != nil {
		fmt.Println(p.name)
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	// dopplerEndpoint, err := cli.DopplerEndpoint()
	// if err != nil {
	// 	p.ui.Failed(err.Error())
	// }

	// url, err := url.Parse(dopplerEndpoint)
	// if err != nil {
	// 	p.ui.Failed(err.Error())
	// }

	token, err := cli.AccessToken()
	if err != nil {
		p.ui.Failed(err.Error())
	}

	client := loggregator.NewRLPGatewayClient(
		"http://log-stream.donaldduck.garden-dev.cf-app.com",
		loggregator.WithRLPGatewayClientLogger(log.New(os.Stderr, "", log.LstdFlags)),
		loggregator.WithRLPGatewayHTTPClient(&tokenAttacher{token: token}),
	)

	stream := client.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{
		Selectors: []*loggregator_v2.Selector{
			{
				SourceId: app.Guid,
				Message:  &loggregator_v2.Selector_Gauge{},
			},
		},
	})

	for {
		for _, e := range stream() {
			usageMetric, ok := NewUsageMetricFromGaugeMetric(e.GetGauge().GetMetrics())
			if !ok {
				continue
			}

			fmt.Printf("CPU usage for %s: %.2f%%\n", p.name, usageMetric.CPUUsage()*100)
		}
	}
}

type UsageMetric struct {
	AbsoluteUsage       float64
	AbsoluteEntitlement float64
	ContainerAge        float64
}

func NewUsageMetricFromGaugeMetric(metric map[string]*loggregator_v2.GaugeValue) (UsageMetric, bool) {
	absoluteUsage := metric["absolute_usage"]
	absoluteEntitlement := metric["absolute_entitlement"]
	containerAge := metric["container_age"]

	if absoluteUsage == nil {
		return UsageMetric{}, false
	}

	if absoluteEntitlement == nil {
		return UsageMetric{}, false
	}

	if containerAge == nil {
		return UsageMetric{}, false
	}

	return UsageMetric{
		AbsoluteUsage:       absoluteUsage.Value,
		AbsoluteEntitlement: absoluteEntitlement.Value,
		ContainerAge:        containerAge.Value,
	}, true
}

func (m UsageMetric) CPUUsage() float64 {
	return m.AbsoluteUsage / m.AbsoluteEntitlement
}

type tokenAttacher struct {
	token string
}

func (a *tokenAttacher) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.token)
	return http.DefaultClient.Do(req)
}

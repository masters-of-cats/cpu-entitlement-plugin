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
	"github.com/masters-of-cats/cpu-entitlement-plugin/usagemetric"
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
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	token, err := cli.AccessToken()
	if err != nil {
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	client := loggregator.NewRLPGatewayClient(
		"http://log-stream.donaldduck.garden-dev.cf-app.com",
		loggregator.WithRLPGatewayClientLogger(log.New(os.Stderr, "", log.LstdFlags)),
		loggregator.WithRLPGatewayHTTPClient(authenticatedBy(token)),
	)

	stream := client.Stream(context.Background(), streamRequest(app.Guid))

	for {
		for _, e := range stream() {
			usageMetric, ok := usagemetric.FromGaugeMetric(e.GetGauge().GetMetrics())
			if !ok {
				continue
			}

			fmt.Printf("CPU usage for %s: %.2f%%\n", p.name, usageMetric.CPUUsage()*100)
		}
	}
}

func streamRequest(sourceID string) *loggregator_v2.EgressBatchRequest {
	return &loggregator_v2.EgressBatchRequest{
		Selectors: []*loggregator_v2.Selector{
			{
				SourceId: sourceID,
				Message:  &loggregator_v2.Selector_Gauge{},
			},
		},
	}
}

func authenticatedBy(token string) *authClient {
	return &authClient{token: token}
}

type authClient struct {
	token string
}

func (a *authClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.token)
	return http.DefaultClient.Do(req)
}

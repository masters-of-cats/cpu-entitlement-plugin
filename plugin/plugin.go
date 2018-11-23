package plugin

import (
	"net/url"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cpu-entitlement-plugin/logstreamer"
)

type CPUEntitlementPlugin struct {
	ui          terminal.UI
	logStreamer logstreamer.LogStreamer
}

func New(logStreamer logstreamer.LogStreamer, ui terminal.UI) *CPUEntitlementPlugin {
	return &CPUEntitlementPlugin{
		ui:          ui,
		logStreamer: logStreamer,
	}
}

func (p *CPUEntitlementPlugin) Run(cli plugin.CliConnection, args []string) {
	if len(args) != 2 {
		p.ui.Failed("Usage: `cf cpu-entitlement APP_NAME`")
		os.Exit(1)
	}

	appName := args[1]

	app, err := cli.GetApp(appName)
	if err != nil {
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	token, err := cli.AccessToken()
	if err != nil {
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	apiURL, err := cli.ApiEndpoint()
	if err != nil {
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	logStreamURL, err := buildLogStreamURL(apiURL)
	if err != nil {
		p.ui.Failed(err.Error())
		os.Exit(1)
	}

	usageMetricsStream := p.logStreamer.Stream(logStreamURL, token, app.Guid)
	for usageMetric := range usageMetricsStream {
		p.ui.Say("CPU usage for %s: %.2f%%\n", appName, usageMetric.CPUUsage()*100)
	}
}

func buildLogStreamURL(apiURL string) (string, error) {
	logStreamURL, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}

	logStreamURL.Scheme = "http"
	logStreamURL.Host = strings.Replace(logStreamURL.Host, "api", "log-stream", 1)

	return logStreamURL.String(), nil
}

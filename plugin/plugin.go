package plugin

import (
	"net/url"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cpu-entitlement-plugin/logstreamer"
)

type CPUEntitlementPlugin struct{}

func New() *CPUEntitlementPlugin {
	return &CPUEntitlementPlugin{}
}

func (p *CPUEntitlementPlugin) Run(cli plugin.CliConnection, args []string) {
	traceLogger := trace.NewLogger(os.Stdout, true, os.Getenv("CF_TRACE"), "")
	ui := terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), traceLogger)

	if len(args) != 2 {
		ui.Failed("Usage: `cf cpu-entitlement APP_NAME`")
		os.Exit(1)
	}

	appName := args[1]

	app, err := cli.GetApp(appName)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	token, err := cli.AccessToken()
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	apiURL, err := cli.ApiEndpoint()
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	logStreamURL, err := buildLogStreamURL(apiURL)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}

	logStreamer := logstreamer.New(logStreamURL, token)

	usageMetricsStream := logStreamer.Stream(app.Guid)
	for usageMetric := range usageMetricsStream {
		ui.Say("CPU usage for %s: %.2f%%", appName, usageMetric.CPUUsage()*100)
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

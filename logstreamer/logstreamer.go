package logstreamer

import (
	"context"
	"log"
	"net/http"
	"os"

	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/cpu-entitlement-plugin/usagemetric"
)

type LogStreamer struct {
}

func New() LogStreamer {
	return LogStreamer{}
}

func (s LogStreamer) Stream(logStreamURL, token, appGuid string) chan usagemetric.UsageMetric {
	client := loggregator.NewRLPGatewayClient(
		logStreamURL,
		loggregator.WithRLPGatewayClientLogger(log.New(os.Stderr, "", log.LstdFlags)),
		loggregator.WithRLPGatewayHTTPClient(authenticatedBy(token)),
	)

	stream := client.Stream(context.Background(), streamRequest(appGuid))

	var usageMetricsStream = make(chan usagemetric.UsageMetric)
	go func() {
		for {
			for _, envelope := range stream() {
				usageMetric, ok := usagemetric.FromGaugeMetric(envelope.GetGauge().GetMetrics())
				if !ok {
					continue
				}

				usageMetricsStream <- usageMetric
			}
		}
	}()

	return usageMetricsStream
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

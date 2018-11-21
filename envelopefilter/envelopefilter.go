package envelopefilter

import (
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/cloudfoundry/sonde-go/events"
)

func Filter(in <-chan *events.Envelope, out chan<- *events.Envelope, filters ...func(*events.Envelope) bool) {
	for {
		envelope := <-in
		for _, filter := range filters {
			if !filter(envelope) {
				continue
			}
		}
		out <- envelope
	}
}

func ValueMetric(envelope *events.Envelope) bool {
	return *envelope.EventType == events.Envelope_ValueMetric
}

func OfApp(app plugin_models.GetAppModel) func(*events.Envelope) bool {
	return func(e *events.Envelope) bool {
		return e.Tags["source_id"] == app.Guid
	}
}

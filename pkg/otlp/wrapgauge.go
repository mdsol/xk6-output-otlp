package otlp

import (
	k6m "go.k6.io/k6/metrics"
	om "go.opentelemetry.io/otel/metric"
)

func newGaugeWrapper(id int, name string) (Wrapper, error) {
	fmetric, err := meter.Float64Gauge(name)
	if err != nil {
		return nil, err
	}
	return &floatGaugeWrapper{
		id:     id,
		metric: fmetric,
	}, nil
}

type floatGaugeWrapper struct {
	id     int
	metric om.Float64Gauge
}

func (w *floatGaugeWrapper) Record(s *k6m.Sample) {
	attrs := om.WithAttributes(attributes(s.TimeSeries.Tags)...)
	w.metric.Record(params.ctx, s.Value, attrs)
}

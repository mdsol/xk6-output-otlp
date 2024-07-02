package otlp

import (
	"math"

	k6m "go.k6.io/k6/metrics"
	om "go.opentelemetry.io/otel/metric"
)

func newRateWrapper(id int, name string) (Wrapper, error) {
	metricTotal, err := meter.Int64Counter(name + "_total")
	if err != nil {
		return nil, err
	}
	metricSuccess, err := meter.Int64Counter(name + "_success")
	if err != nil {
		return nil, err
	}

	metricRate, err := meter.Float64Gauge(name + "_rate")
	if err != nil {
		return nil, err
	}

	return &rateWrapper{
		id:            id,
		metricTotal:   metricTotal,
		metricSuccess: metricSuccess,
		metricRate:    metricRate,
		rate:          &rate{},
	}, nil
}

type rateWrapper struct {
	id            int
	metricTotal   om.Int64Counter
	metricSuccess om.Int64Counter
	metricRate    om.Float64Gauge
	rate          *rate
}

func (w *rateWrapper) Record(s *k6m.Sample) {
	attrs := om.WithAttributes(attributes(s.TimeSeries.Tags)...)

	w.metricRate.Record(params.ctx, w.rate.combine(s))
	w.metricTotal.Add(params.ctx, 1, attrs)
	w.metricSuccess.Add(params.ctx, int64(math.Floor(s.Value)), attrs)
}

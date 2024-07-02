package otlp

import (
	"math"

	k6m "go.k6.io/k6/metrics"
	om "go.opentelemetry.io/otel/metric"
)

func newCounterWrapper(id int, name string, isFloat bool) (Wrapper, error) {
	if isFloat {
		fmetric, err := meter.Float64Counter(name)
		if err != nil {
			return nil, err
		}
		return &floatCounterWrapper{
			id:     id,
			metric: fmetric,
		}, nil
	}

	imetric, err := meter.Int64Counter(name)
	if err != nil {
		return nil, err
	}
	return &intCounterWrapper{
		id:     id,
		metric: imetric,
	}, nil
}

type intCounterWrapper struct {
	id     int
	metric om.Int64Counter
}

type floatCounterWrapper struct {
	id     int
	metric om.Float64Counter
}

func (w *intCounterWrapper) Record(s *k6m.Sample) {
	attrs := om.WithAttributes(attributes(s.TimeSeries.Tags)...)
	w.metric.Add(params.ctx, int64(math.Floor(s.Value)), attrs)
}

func (w *floatCounterWrapper) Record(s *k6m.Sample) {
	attrs := om.WithAttributes(attributes(s.TimeSeries.Tags)...)
	w.metric.Add(params.ctx, s.Value, attrs)
}

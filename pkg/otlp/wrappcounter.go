package otlp

import (
	"math"

	k6m "go.k6.io/k6/metrics"
	om "go.opentelemetry.io/otel/metric"
)

func newCounterWrapper(id int, name string, isFloat bool) (Wrapper, error) {
	if isFloat {
		fmetric, err := meter.Float64Counter(name, om.WithUnit("ms"))
		if err != nil {
			return nil, err
		}
		return &floatCounterWrapper{
			id:     id,
			metric: fmetric,
		}, nil
	}

	imetric, err := meter.Int64Counter(name, om.WithUnit("1"))
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
	attrs := attributes(s.TimeSeries.Tags)
	w.metric.Add(params.ctx, int64(math.Floor(s.Value)), om.WithAttributes(attrs...))
}

func (w *floatCounterWrapper) Record(s *k6m.Sample) {
	attrs := attributes(s.TimeSeries.Tags)
	w.metric.Add(params.ctx, s.Value, om.WithAttributes(attrs...))
}

package otlp

import (
	"math"

	k6m "go.k6.io/k6/metrics"
	om "go.opentelemetry.io/otel/metric"
)

func newHistogramWrapper(id int, name string, isFloat bool) (Wrapper, error) {
	if isFloat {
		fmetric, err := meter.Float64Histogram(name)
		if err != nil {
			return nil, err
		}

		return &floatHistogramWrapper{id: id, metric: fmetric}, nil
	}

	imetric, err := meter.Int64Histogram(name)
	if err != nil {
		return nil, err
	}

	return &intHistogramWrapper{id: id, metric: imetric}, nil
}

type floatHistogramWrapper struct {
	id     int
	metric om.Float64Histogram
}

type intHistogramWrapper struct {
	id     int
	metric om.Int64Histogram
}

func (w *floatHistogramWrapper) Record(s *k6m.Sample) {
	attrs := attributes(s.TimeSeries.Tags)
	w.metric.Record(params.ctx, s.Value, om.WithAttributes(attrs...))
}

func (w *intHistogramWrapper) Record(s *k6m.Sample) {
	attrs := attributes(s.TimeSeries.Tags)
	w.metric.Record(params.ctx, int64(math.Floor(s.Value)), om.WithAttributes(attrs...))
}

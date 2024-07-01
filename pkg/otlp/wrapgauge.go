package otlp

import (
	"math"
	"strings"

	k6m "go.k6.io/k6/metrics"
	"go.opentelemetry.io/otel/attribute"
	om "go.opentelemetry.io/otel/metric"
)

func newGaugeWrapper(id int, name string, isFloat bool, isTrend bool) (Wrapper, error) {
	if isFloat {
		fmetric, err := meter.Float64Gauge(name)
		if err != nil {
			return nil, err
		}
		return &floatGaugeWrapper{
			id:      id,
			metric:  fmetric,
			isTrend: isTrend,
		}, nil
	}

	imetric, err := meter.Int64Gauge(name)
	if err != nil {
		return nil, err
	}

	return &intGaugeWrapper{
		id:     id,
		metric: imetric,
	}, nil
}

type floatGaugeWrapper struct {
	id      int
	metric  om.Float64Gauge
	isTrend bool
}
type intGaugeWrapper struct {
	id     int
	metric om.Int64Gauge
}

func (w *floatGaugeWrapper) Record(s *k6m.Sample) {
	attrs := attributes(s.TimeSeries.Tags)

	if w.isTrend {
		for k, v := range s.Metric.Sink.Format(0) {
			temp := strings.Replace(strings.Replace(k, "(", "", 1), ")", "", 1)
			statPair := attribute.KeyValue{Key: "stat", Value: attribute.StringValue(temp)}
			extAttrs := []attribute.KeyValue{statPair}
			extAttrs = append(extAttrs, attrs...)
			w.metric.Record(params.ctx, v, om.WithAttributes(extAttrs...))
		}
	}

	w.metric.Record(params.ctx, s.Value, om.WithAttributes(attrs...))
}

func (w *intGaugeWrapper) Record(s *k6m.Sample) {
	attrs := attributes(s.TimeSeries.Tags)
	w.metric.Record(params.ctx, int64(math.Floor(s.Value)), om.WithAttributes(attrs...))
}

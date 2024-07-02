package otlp

import (
	"math"
	"strings"

	k6m "go.k6.io/k6/metrics"
	"go.opentelemetry.io/otel/attribute"
	om "go.opentelemetry.io/otel/metric"
)

func newTrendWrapper(id int, name string, isTime bool) (Wrapper, error) {
	createHistogram := func(name string, isFloat bool) (*om.Float64Histogram, *om.Int64Histogram, error) {
		if isFloat {
			fmetric, err := meter.Float64Histogram(name)
			if err != nil {
				return nil, nil, err
			}

			return &fmetric, nil, nil
		}

		imetric, err := meter.Int64Histogram(name)
		if err != nil {
			return nil, nil, err
		}

		return nil, &imetric, nil
	}

	fmetric, imetric, err := createHistogram(name, isTime)
	if err != nil {
		return nil, err
	}

	precalc, err := meter.Float64Gauge(name + "_stat")
	if err != nil {
		return nil, err
	}

	return &trendWrapper{
		id:          id,
		isFloat:     isTime,
		floatMetric: fmetric,
		intMetric:   imetric,
		preCalcStat: &precalc,
	}, nil
}

type trendWrapper struct {
	id          int
	isFloat     bool
	floatMetric *om.Float64Histogram
	intMetric   *om.Int64Histogram
	preCalcStat *om.Float64Gauge
}

func (w *trendWrapper) Record(s *k6m.Sample) {
	attrs := attributes(s.TimeSeries.Tags)
	w.recordPreCalculated(s, attrs)

	if w.isFloat {
		(*w.floatMetric).Record(params.ctx, s.Value, om.WithAttributes(attrs...))
		return
	}

	(*w.intMetric).Record(params.ctx, int64(math.Round(s.Value)), om.WithAttributes(attrs...))
}

func (w *trendWrapper) recordPreCalculated(s *k6m.Sample, attrs []attribute.KeyValue) {
	for k, v := range s.Metric.Sink.Format(0) {
		temp := strings.Replace(strings.Replace(k, "(", "", 1), ")", "", 1)
		if len(temp) > 0 {
			statPair := attribute.KeyValue{Key: "stat", Value: attribute.StringValue(temp)}
			extAttrs := []attribute.KeyValue{statPair}
			extAttrs = append(extAttrs, attrs...)
			(*w.preCalcStat).Record(params.ctx, v, om.WithAttributes(extAttrs...))
		}
	}
}

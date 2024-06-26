package otlp

import (
	"fmt"

	k6m "go.k6.io/k6/metrics"
	"go.opentelemetry.io/otel/attribute"
)

const DefaultMetricPrefix = "k6_"

var (
	wrapperCount = 0
)

type Wrapper interface {
	Record(*k6m.Sample)
}

func NewWrapper(s k6m.Sample) (Wrapper, error) {
	var (
		err    error
		retval Wrapper
	)
	wrapperCount++

	switch s.Metric.Type {
	case k6m.Counter:
		retval, err = newCounterWrapper(wrapperCount, metricName(&s), s.Metric.Contains == k6m.Time)
	case k6m.Gauge:
		fallthrough
	case k6m.Rate:
		if params.rateConversion == "gauge" {
			retval, err = newGaugeWrapper(wrapperCount, metricName(&s), true, false)
		} else {
			retval, err = newCounterWrapper(wrapperCount, metricName(&s), true)
		}
	case k6m.Trend:
		if s.Metric.Contains == k6m.Time {
			if params.trendConversion == "histogram" {
				retval, err = newHistogramWrapper(wrapperCount, metricName(&s), true)
			} else {
				retval, err = newGaugeWrapper(wrapperCount, metricName(&s), true, true)
			}
		} else {
			if params.trendConversion == "histogram" {
				retval, err = newHistogramWrapper(wrapperCount, metricName(&s), false)
			} else {
				retval, err = newGaugeWrapper(wrapperCount, metricName(&s), false, true)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return retval, nil
}
func attributes(tags *k6m.TagSet) []attribute.KeyValue {
	retval := make([]attribute.KeyValue, 0, len(sessionAttrs)+len(tags.Map())-1)
	retval = append(retval, sessionAttrs...)

	for key, value := range tags.Map() {
		if key != "__name__" {
			retval = append(retval, attribute.String(key, value))
		}
	}

	return retval
}

func metricName(s *k6m.Sample) string {
	return fmt.Sprintf("%s%s", DefaultMetricPrefix, s.Metric.Name)
}

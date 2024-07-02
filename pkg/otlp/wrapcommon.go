package otlp

import (
	"fmt"

	"github.com/sirupsen/logrus"
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

func NewWrapper(logger logrus.FieldLogger, s k6m.Sample) (Wrapper, error) {
	var (
		err    error
		retval Wrapper
	)
	wrapperCount++
	name := metricName(&s)

	switch s.Metric.Type {
	case k6m.Counter:
		logger.Debugf("New OTLP for %s [counter]", name)
		retval, err = newCounterWrapper(wrapperCount, name, s.Metric.Contains == k6m.Time)
	case k6m.Gauge:
		logger.Debugf("New OTLP for %s [gauge]", name)
		retval, err = newGaugeWrapper(wrapperCount, name)
	case k6m.Rate:
		logger.Debugf("New OTLP for %s [rate]", name)
		retval, err = newRateWrapper(wrapperCount, name)
	case k6m.Trend:
		logger.Debugf("New OTLP for %s [trend]", name)
		retval, err = newTrendWrapper(wrapperCount, name, s.Metric.Contains == k6m.Time)
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

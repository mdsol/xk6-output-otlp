package otlp

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/sirupsen/logrus"
	k6m "go.k6.io/k6/metrics"
	"go.opentelemetry.io/otel/attribute"
	om "go.opentelemetry.io/otel/metric"
)

var wrapperCount = 1

type wrapper interface {
	apply(context.Context, logrus.FieldLogger, *k6m.Sample) error
}

func newWrapper(_ logrus.FieldLogger, s k6m.Sample, conf *Config, ids *idAttrs) (wrapper, error) {
	var err error
	retval := &omWrapper{script: conf.Script, ids: ids}

	switch s.Metric.Type {
	case k6m.Counter:
		if s.Metric.Contains == k6m.Time {
			retval.metric, err = meter.Float64Counter(s.Metric.Name, om.WithUnit("ms"))
		} else {
			retval.metric, err = meter.Int64Counter(s.Metric.Name, om.WithUnit("1"))
		}
	case k6m.Gauge:
		fallthrough
	case k6m.Rate:
		if conf.RateConversion.String == "gauge" {
			retval.metric, err = meter.Float64Gauge(s.Metric.Name)
		} else {
			retval.metric, err = meter.Float64Counter(s.Metric.Name)
		}
	case k6m.Trend:
		if s.Metric.Contains == k6m.Time {
			if conf.TrendConversion.String == "histogram" {
				retval.metric, err = meter.Float64Histogram(s.Metric.Name, om.WithUnit("ms"))
			} else {
				retval.metric, err = meter.Float64Gauge(s.Metric.Name, om.WithUnit("ms"))
			}
		} else {
			if conf.TrendConversion.String == "histogram" {
				retval.metric, err = meter.Int64Histogram(s.Metric.Name, om.WithUnit("1"))
			} else {
				retval.metric, err = meter.Int64Gauge(s.Metric.Name, om.WithUnit("1"))
			}
		}
	}

	if err != nil {
		return nil, err
	}

	wrapperCount++
	retval.id = wrapperCount
	return retval, nil
}

type omWrapper struct {
	id     int
	metric any
	script string
	ids    *idAttrs
}

func (w *omWrapper) apply(ctx context.Context, _ logrus.FieldLogger, s *k6m.Sample) error {
	attrs := w.attributes(s.TimeSeries.Tags)
	var (
		ok       bool
		fcounter om.Float64Counter
		fgauge   om.Float64Gauge
		icounter om.Int64Counter
		ihist    om.Int64Histogram
	)
	switch s.Metric.Type {
	case k6m.Counter:
		if s.Metric.Contains == k6m.Time {
			fcounter, ok = w.metric.(om.Float64Counter)
			if ok {
				fcounter.Add(ctx, s.Value, om.WithAttributes(attrs...))
			}
		} else {
			icounter, ok = w.metric.(om.Int64Counter)
			if ok {
				icounter.Add(ctx, int64(math.Floor(s.Value)), om.WithAttributes(attrs...))
			}
		}
	case k6m.Gauge:
		fallthrough
	case k6m.Rate:
		fgauge, ok = w.metric.(om.Float64Gauge)
		if ok {
			fgauge.Record(ctx, s.Value, om.WithAttributes(attrs...))
		} else {
			fcounter, ok := w.metric.(om.Float64Counter)
			if ok {
				fcounter.Add(ctx, s.Value, om.WithAttributes(attrs...))
			}
		}
	case k6m.Trend:
		if s.Metric.Contains == k6m.Time {
			ok = recordFloatTrend(s, w, attrs)
		} else {
			ihist, ok = w.metric.(om.Int64Histogram)
			if ok {
				ihist.Record(ctx, int64(math.Floor(s.Value)), om.WithAttributes(attrs...))
			}
		}
	}

	if !ok {
		return fmt.Errorf("OTel metric %s is not found", s.Metric.Name)
	}

	return nil
}

func recordFloatTrend(s *k6m.Sample, w *omWrapper, attrs []attribute.KeyValue) bool {
	// Gauges scenario
	fg, ok := w.metric.(om.Float64Gauge)
	if ok {
		for k, v := range s.Metric.Sink.Format(0) {
			temp := strings.Replace(strings.Replace(k, "(", "", 1), ")", "", 1)
			statPair := attribute.KeyValue{Key: "stat", Value: attribute.StringValue(temp)}
			extAttrs := []attribute.KeyValue{statPair}
			extAttrs = append(extAttrs, attrs...)
			fg.Record(ctx, v, om.WithAttributes(extAttrs...))
		}
	}

	// Histogram scenario
	fh, ok := w.metric.(om.Float64Histogram)
	if ok {
		fh.Record(ctx, s.Value, om.WithAttributes(attrs...))
	}
	return ok
}

func (w *omWrapper) attributes(tags *k6m.TagSet) []attribute.KeyValue {
	retval := []attribute.KeyValue{
		attribute.String("provider", "k6"),
		attribute.String("script", w.script),
	}

	for key, value := range tags.Map() {
		if key != "__name__" {
			retval = append(retval, attribute.String(key, value))
		}
	}

	if w.ids != nil {
		retval = append(retval,
			attribute.String("provider_id", w.ids.providerID),
			attribute.Int("run_id", int(w.ids.runID)),
		)
	}

	return retval
}

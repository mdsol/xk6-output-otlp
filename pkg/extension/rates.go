package extension

import (
	k6m "go.k6.io/k6/metrics"
)

type rate struct {
	first *k6m.Sample
	sum   float64
	count float64
}

func (r *rate) value() float64 {
	return r.sum / r.count
}

func (r *rate) combine(sample k6m.Sample) float64 {
	r.count += 1.0
	r.sum += sample.Value
	return r.value()
}

func (r *rate) Result(rateConversion string) []k6m.Sample {
	if rateConversion == "counters" {
		return []k6m.Sample{
			{
				TimeSeries: k6m.TimeSeries{
					Metric: &k6m.Metric{
						Name:     r.first.Metric.Name,
						Type:     k6m.Counter,
						Contains: k6m.Default,
						Sink:     &k6m.CounterSink{},
					},
					Tags: r.first.Tags.With("value", "1"),
				},
				Metadata: r.first.Metadata,
				Time:     r.first.Time,
				Value:    r.sum,
			},
			{
				TimeSeries: k6m.TimeSeries{
					Metric: &k6m.Metric{
						Name:     r.first.Metric.Name,
						Type:     k6m.Counter,
						Contains: k6m.Default,
						Sink:     &k6m.CounterSink{},
					},
					Tags: r.first.Tags.With("value", "0"),
				},
				Metadata: r.first.Metadata,
				Time:     r.first.Time,
				Value:    r.count - r.sum,
			},
		}
	}

	return []k6m.Sample{
		{
			TimeSeries: k6m.TimeSeries{
				Metric: &k6m.Metric{
					Name:     r.first.Metric.Name + "_rate",
					Type:     k6m.Gauge,
					Contains: k6m.Default,
					Sink:     &k6m.GaugeSink{},
				},
				Tags: r.first.Tags,
			},
			Metadata: r.first.Metadata,
			Time:     r.first.Time,
			Value:    r.value(),
		},
	}
}

func newRate(sample *k6m.Sample) *rate {
	return &rate{
		first: sample,
		count: 1.0,
		sum:   sample.Value,
	}
}

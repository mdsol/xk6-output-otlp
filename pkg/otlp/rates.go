package otlp

import (
	hash "crypto/sha256"
	"encoding/hex"

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

func joinRates(samples []k6m.Sample, rateConversion string) []k6m.Sample {
	retval := []k6m.Sample{}
	rates := map[string]*rate{}

	for i := 0; i < len(samples); i++ {
		if samples[i].Metric.Type != k6m.Rate {
			retval = append(retval, samples[i])
			continue
		}

		k, err := key(samples[i].TimeSeries)
		if err != nil {
			logger.WithError(err).Debug("key serialization")
			continue
		}

		existing, ok := rates[k]
		if !ok {
			rates[k] = newRate(&samples[i])
			continue
		}

		_ = existing.combine(samples[i])
	}

	for k, v := range rates {
		retval = append(retval, v.Result(rateConversion)...)
		logger.WithField("key", k).WithField("value", v.value()).Debug("RATE")
	}

	logger.Debugf("SAMPLES before %d after %d", len(samples), len(retval))
	return retval
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

func flatten(containers []k6m.SampleContainer) []k6m.Sample {
	retval := []k6m.Sample{}

	for i := 0; i < len(containers); i++ {
		retval = append(retval, containers[i].GetSamples()...)
	}

	return retval
}

func key(ts k6m.TimeSeries) (string, error) {
	data, err := ts.Tags.MarshalJSON()
	if err != nil {
		return ts.Metric.Name, err
	}

	var key = hash.Sum256([]byte(ts.Metric.Name + "|" + string(data)))

	return hex.EncodeToString(key[:]), nil
}

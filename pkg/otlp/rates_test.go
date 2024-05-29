package otlp

import (
	"testing"
	"time"

	k6m "go.k6.io/k6/metrics"
	"gotest.tools/v3/assert"
)

func TestRate(t *testing.T) {
	sample := &k6m.Sample{
		TimeSeries: k6m.TimeSeries{
			Metric: &k6m.Metric{
				Name:     "test_rate",
				Type:     k6m.Gauge,
				Contains: k6m.Default,
				Sink:     &k6m.GaugeSink{},
			},
			Tags: &k6m.TagSet{},
		},
		Metadata: map[string]string{},
		Time:     time.Now(),
		Value:    1,
	}

	r := newRate(sample)

	r.combine(k6m.Sample{
		Value: 0,
	})

	r.combine(k6m.Sample{
		Value: 0,
	})

	r.combine(k6m.Sample{
		Value: 1,
	})

	_ = r.Result()

	assert.Equal(t, r.value, 0.5)
}

package otlp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k6m "go.k6.io/k6/metrics"
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

	assert.Equal(t, r.value, 1.0)

	r.combine(k6m.Sample{
		Value: 0,
	})
	assert.Equal(t, r.value, 0.5)

	r.combine(k6m.Sample{
		Value: 0,
	})
	assert.Equal(t, r.value, 1.0/3)

	r.combine(k6m.Sample{
		Value: 1,
	})
	assert.Equal(t, r.value, 0.5)

	res := r.Result()
	require.NotNil(t, res)
	assert.Equal(t, res.Value, r.value)

}

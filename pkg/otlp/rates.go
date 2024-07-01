package otlp

import (
	k6m "go.k6.io/k6/metrics"
)

type rate struct {
	sum   float64
	count float64
}

func (r *rate) value() float64 {
	if r.count == 0 {
		return 0
	}
	return r.sum / r.count
}

func (r *rate) combine(sample *k6m.Sample) float64 {
	r.count += 1.0
	r.sum += sample.Value
	return r.value()
}

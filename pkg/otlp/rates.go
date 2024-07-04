package otlp

import (
	k6m "go.k6.io/k6/metrics"
)

type rate struct {
	sum     float64
	counter float64
}

func (r *rate) value() float64 {
	if r.counter == 0 {
		return 0
	}
	return r.sum / r.counter
}

func (r *rate) combine(sample *k6m.Sample) float64 {
	r.counter += 1.0
	r.sum += sample.Value
	return r.value()
}

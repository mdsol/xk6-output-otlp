package otlp

import (
	"time"
)

type Config struct {
	Script          string
	TrendConversion string
	RateConversion  string
	Timeout         time.Duration
	PushInterval    time.Duration
	UseIDs          bool
}

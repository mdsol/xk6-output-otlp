package otlp

import (
	"time"
)

type Config struct {
	Script       string
	Timeout      time.Duration
	PushInterval time.Duration
	UseIDs       bool
}

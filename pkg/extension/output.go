package extension

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mdsol/xk6-output-otlp/pkg/otlp"

	"github.com/sirupsen/logrus"

	k6m "go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

var (
	logger logrus.FieldLogger
)

type Output struct {
	output.SampleBuffer

	ctx             context.Context
	logger          logrus.FieldLogger
	config          *Config
	now             func() time.Time
	periodicFlusher *output.PeriodicFlusher

	otelMetrics sync.Map

	down *atomic.Bool
}

func New(params output.Params) (*Output, error) {
	logger = params.Logger
	c, err := parseJSON(params.JSONConfig)

	params.Logger.
		WithField("script_path", params.ScriptPath.Path).
		WithField("json_config", c).WithError(err).
		Debug("Params")

	conf, err := joinConfig(params.JSONConfig, params.Environment, params.Logger)
	if err != nil {
		return nil, err
	}

	if params.ScriptPath != nil {
		seg := strings.Split(params.ScriptPath.Path, "/")
		conf.Script = seg[len(seg)-1]
	}

	fields := logrus.Fields{
		"endpoint":     conf.ServerURL.String,
		"gzip":         conf.GZip.Bool,
		"pushInterval": conf.PushInterval,
		"insecure":     conf.Insecure.Bool,
		"timeout":      conf.Timeout,
		"headers":      conf.Headers,
		"add_ids":      conf.AddIDAttributes,
	}
	params.Logger.WithFields(fields).Debug("OTEL Config")

	expc, wrapc, err := conf.PartualConfigs()
	if err != nil {
		return nil, err
	}

	o := &Output{
		ctx:    context.Background(),
		logger: params.Logger,
		config: &conf,
		now:    time.Now,
	}

	o.down = &atomic.Bool{}

	err = otlp.Init(wrapc, expc, params.Logger)

	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Output) Description() string {
	return fmt.Sprintf("OTLP (%s)", o.config.ServerURL.String)
}

func (o *Output) Start() error {
	var err error

	d := o.config.PushInterval.TimeDuration()
	periodicFlusher, err := output.NewPeriodicFlusher(d, o.flush)
	if err != nil {
		return err
	}
	o.periodicFlusher = periodicFlusher

	o.logger.WithField("flush_time", d).Debug("Output initialized")
	return nil
}

func (o *Output) Stop() error {
	defer o.logger.Debug("Output stopped")

	o.logger.Debug("Stopping the output")
	o.periodicFlusher.Stop()

	err := otlp.Shutdown()
	if err != nil {
		o.logger.Error(err)
	}

	return nil
}

func (o *Output) flush() {
	if o.down.Load() {
		return
	}

	samples := o.GetBufferedSamples()

	if len(samples) < 1 {
		o.logger.Debug("No buffered samples, skip exporting")
		return
	}

	o.applyMetrics(samples)
}

func (o *Output) applyMetrics(samplesContainers []k6m.SampleContainer) {
	var err error

	input := flatten(samplesContainers)
	if o.config.RateConversion.String == "gauge" {
		input = joinRates(input, o.config.RateConversion.String)
	}

	for _, s := range input {
		w, found := o.otelMetrics.Load(s.Metric.Name)
		if !found {
			w, err = otlp.NewWrapper(s)
			if err != nil {
				o.logger.Errorf("Unable to wrap %s:[%v] metric\n", s.Metric.Name, s.Metric.Type)
				continue
			}

			o.otelMetrics.Store(s.Metric.Name, w)
		}

		err = w.(otlp.Wrapper).Record(o.ctx, o.logger, &s)
		if err != nil {
			o.logger.WithError(err).Errorf("Unable to apply %s:[%v] metric", s.Metric.Name, s.Metric.Type)
		}
	}
}

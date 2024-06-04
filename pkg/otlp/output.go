package otlp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mdsol/xk6-output-otlp/pkg/exporter"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"

	omhttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	om "go.opentelemetry.io/otel/metric"
	ocm "go.opentelemetry.io/otel/sdk/metric"
	ores "go.opentelemetry.io/otel/sdk/resource"

	k6m "go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

var (
	reader   *ocm.PeriodicReader
	provider *ocm.MeterProvider
	meter    om.Meter
	ctx      = context.Background()
	logger   logrus.FieldLogger
)

type Output struct {
	output.SampleBuffer

	ctx             context.Context
	logger          logrus.FieldLogger
	config          *Config
	now             func() time.Time
	periodicFlusher *output.PeriodicFlusher
	tsdb            map[k6m.TimeSeries]wrapper

	client   *omhttp.Exporter
	runCount om.Int64Counter
	down     *atomic.Bool
	ids      *idAttrs
}

func New(params output.Params) (*Output, error) {
	logger = params.Logger
	c, err := parseJSON(params.JSONConfig)

	params.Logger.
		WithField("script_path", params.ScriptPath.Path).
		WithField("json_config", c).WithError(err).
		Debug("Params")

	conf, err := joinConfig(params.JSONConfig, params.Environment)
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

	expconf, err := conf.ExporterConfig()
	if err != nil {
		return nil, err
	}

	exp, err := exporter.New(expconf)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize the OTLP exporter: %s", err.Error())
	}

	var ids *idAttrs

	if conf.AddIDAttributes.Bool {
		ids, err = newIdentities()
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize ID attributes: %s", err.Error())
		}
	}

	o := &Output{
		ctx:    context.Background(),
		logger: params.Logger,
		client: exp,
		config: &conf,
		now:    time.Now,
		tsdb:   make(map[k6m.TimeSeries]wrapper),
		ids:    ids,
	}
	reader = ocm.NewPeriodicReader(exp,
		ocm.WithInterval(conf.PushInterval.TimeDuration()),
		ocm.WithTimeout(conf.Timeout.TimeDuration()))

	attrs := []attribute.KeyValue{
		{Key: "provider", Value: attribute.StringValue("k6")},
	}

	if conf.Script != "" {
		attrs = append(attrs, attribute.String("script", conf.Script))
	}

	res := ores.NewSchemaless(attrs...)

	provider = ocm.NewMeterProvider(
		ocm.WithReader(reader),
		ocm.WithResource(res),
	)

	meter = provider.Meter("K6")
	o.down = &atomic.Bool{}
	return o, nil
}

func (o *Output) Description() string {
	return fmt.Sprintf("OTLP (%s)", o.config.ServerURL.String)
}

func (o *Output) Start() error {
	var err error
	o.runCount, err = meter.Int64Counter("flush_metrics", om.WithDescription("The number of times the output flushed"))
	if err != nil {
		log.Fatal(err)
	}

	d := o.config.PushInterval.TimeDuration()
	periodicFlusher, err := output.NewPeriodicFlusher(d, o.flush)
	if err != nil {
		return err
	}
	o.periodicFlusher = periodicFlusher

	o.logger.WithField("flushtime", d).Debug("Output initialized")
	return nil
}

func (o *Output) Stop() error {
	defer o.logger.Debug("Output stopped")

	o.logger.Debug("Stopping the output")
	o.periodicFlusher.Stop()

	err := provider.Shutdown(o.ctx)
	if err != nil {
		o.logger.Error(err)
	}

	_ = o.client.ForceFlush(ctx)
	_ = o.client.Shutdown(o.ctx)

	return nil
}

func (o *Output) flush() {
	if o.down.Load() {
		return
	}

	o.runCount.Add(o.ctx, 1)

	samples := o.GetBufferedSamples()

	if len(samples) < 1 {
		o.logger.Debug("No buffered samples, skip exporting")
		return
	}

	o.applyMetrics(samples)
}

func (o *Output) applyMetrics(samplesContainers []k6m.SampleContainer) {
	var err error

	samples := joinRates(flatten(samplesContainers), o.config.RateConversion.String)

	for _, s := range samples {
		w, found := o.tsdb[s.TimeSeries]
		if !found {
			w, err = newWrapper(o.logger, s, o.config, o.ids)
			if err != nil {
				o.logger.Errorf("Unable to wrap %s:[%v] metric\n", s.Metric.Name, s.Metric.Type)
				continue
			}

			o.tsdb[s.TimeSeries] = w
		}

		err = w.apply(o.ctx, o.logger, &s)
		if err != nil {
			o.logger.WithError(err).Errorf("Unable to apply %s:[%v] metric", s.Metric.Name, s.Metric.Type)
		}
	}
}

package otlp

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"

	om "go.opentelemetry.io/otel/metric"
	ocm "go.opentelemetry.io/otel/sdk/metric"
	ores "go.opentelemetry.io/otel/sdk/resource"

	exp "github.com/mdsol/xk6-output-otlp/pkg/exporter"
)

var (
	params       *wrapParams
	flushCount   om.Int64Counter
	provider     *ocm.MeterProvider
	meter        om.Meter
	sessionAttrs = []attribute.KeyValue{
		attribute.String("provider", "k6"),
	}
)

type wrapParams struct {
	ctx    context.Context
	script string
	log    logrus.FieldLogger
}

func Init(conf *Config, expc *exp.Config, log logrus.FieldLogger) error {
	var err error

	params = &wrapParams{
		ctx:    context.Background(),
		script: conf.Script,
		log:    log,
	}

	exporter, err := exp.New(expc)
	if err != nil {
		return err
	}

	reader := ocm.NewPeriodicReader(exporter, ocm.WithInterval(conf.PushInterval), ocm.WithTimeout(conf.Timeout))
	provider = ocm.NewMeterProvider(ocm.WithReader(reader), ocm.WithResource(ores.NewSchemaless(sessionAttrs...)))
	meter = provider.Meter("K6")

	flushCount, err = meter.Int64Counter(
		fmt.Sprintf("%sflush_metrics", DefaultMetricPrefix),
		om.WithDescription("The number of times the output flushed"))
	if err != nil {
		return err
	}

	if conf.Script != "" {
		sessionAttrs = append(sessionAttrs, attribute.String("script", conf.Script))
	}

	if conf.UseIDs {
		ids, err := newIdentities()
		if err == nil {
			sessionAttrs = append(sessionAttrs,
				attribute.String("provider_id", ids.providerID),
				attribute.Int("run_id", int(ids.runID)))
		}
	}

	return nil
}

func Flush() {
	params.log.Debug("Flushing metrics...")
	flushCount.Add(params.ctx, 1, om.WithAttributes(sessionAttrs...))
}

func Shutdown() error {
	err := provider.Shutdown(params.ctx)
	if err != nil {
		return err
	}

	return nil
}

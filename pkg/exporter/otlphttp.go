package exporter

import (
	"context"

	otlp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
)

func New(conf *Config) (*otlp.Exporter, error) {
	options := []otlp.Option{
		otlp.WithEndpointURL(conf.Endpoint),
		otlp.WithTimeout(conf.Timeout),
	}

	if conf.GzipCompression {
		options = append(options, otlp.WithCompression(otlp.GzipCompression))
	}

	if conf.Insecure {
		options = append(options, otlp.WithInsecure())
	}

	retval, err := otlp.New(context.Background(), options...)
	if err != nil {
		return nil, err
	}

	return retval, nil
}

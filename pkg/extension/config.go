package extension

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mdsol/xk6-output-otlp/pkg/exporter"
	"github.com/mdsol/xk6-output-otlp/pkg/otlp"

	"github.com/sirupsen/logrus"

	"go.k6.io/k6/lib/types"
	"gopkg.in/guregu/null.v3"
)

const (
	defaultServerURL       = "http://localhost:8080/v1/metrics"
	defaultTimeout         = 5 * time.Second
	defaultPushInterval    = 5 * time.Second
	defaultRateConversion  = "counters"
	defaultTrendConversion = "gauges"
)

type Config struct {
	AddIDAttributes null.Bool          `json:"add_id_attributes"`
	GZip            null.Bool          `json:"gzip"`
	Headers         map[string]string  `json:"headers"`
	Insecure        null.Bool          `json:"insecure"`
	PushInterval    types.NullDuration `json:"push_interval"`
	RateConversion  null.String        `json:"rate_conversion"`
	Script          string             `json:"-"`
	ServerURL       null.String        `json:"metrics_url"`
	Timeout         types.NullDuration `json:"timeout"`
	TrendConversion null.String        `json:"trend_conversion"`
}

func NewConfig() Config {
	return Config{
		AddIDAttributes: null.BoolFrom(false),
		GZip:            null.BoolFrom(false),
		Insecure:        null.BoolFrom(true),

		PushInterval:    types.NullDurationFrom(defaultPushInterval),
		ServerURL:       null.StringFrom(defaultServerURL),
		Headers:         make(map[string]string),
		Timeout:         types.NullDurationFrom(defaultTimeout),
		TrendConversion: null.StringFrom(defaultTrendConversion),
	}
}

func (c Config) PartialConfigs() (*exporter.Config, *otlp.Config, error) {
	expc := &exporter.Config{
		Timeout:         defaultTimeout,
		Endpoint:        c.ServerURL.String,
		Headers:         http.Header{},
		GzipCompression: c.GZip.Bool,
		Insecure:        c.Insecure.Bool,
	}

	if len(c.Headers) > 0 {
		for key, val := range c.Headers {
			expc.Headers.Add(key, val)
		}
	}

	err := expc.Validate()
	if err != nil {
		return nil, nil, err
	}

	wrapc := &otlp.Config{
		Script:          c.Script,
		TrendConversion: c.TrendConversion.String,
		RateConversion:  c.RateConversion.String,
		Timeout:         c.Timeout.TimeDuration(),
		PushInterval:    c.PushInterval.TimeDuration(),
		UseIDs:          c.AddIDAttributes.Bool,
	}

	return expc, wrapc, nil
}

// Apply merges applied Config into base.
func (c Config) Apply(applied Config) Config {
	if applied.ServerURL.Valid {
		c.ServerURL = applied.ServerURL
	}

	if applied.PushInterval.Valid {
		c.PushInterval = applied.PushInterval
	}

	if len(applied.Headers) > 0 {
		for k, v := range applied.Headers {
			c.Headers[k] = v
		}
	}

	if applied.Timeout.Valid {
		c.Timeout = applied.Timeout
	}

	if applied.Insecure.Valid {
		c.Insecure = applied.Insecure
	}

	if applied.GZip.Valid {
		c.GZip = applied.GZip
	}

	if applied.AddIDAttributes.Valid {
		c.AddIDAttributes = applied.AddIDAttributes
	}

	if applied.RateConversion.Valid {
		c.RateConversion = applied.RateConversion
	}

	if applied.TrendConversion.Valid {
		c.TrendConversion = applied.TrendConversion
	}

	return c
}

func joinConfig(jsonRawConf json.RawMessage, env map[string]string, log logrus.FieldLogger) (Config, error) {
	result := NewConfig()
	if jsonRawConf != nil {
		jsonConf, err := parseJSON(jsonRawConf)
		if err != nil {
			return result, fmt.Errorf("parse JSON options failed: %w", err)
		}
		log.Debug("json_config", string(jsonRawConf))
		result = result.Apply(jsonConf)
	}

	if len(env) > 0 {
		envConf, err := parseEnvs(env)
		if err != nil {
			return result, fmt.Errorf("parse environment variables options failed: %w", err)
		}
		result = result.Apply(envConf)
	}

	if result.RateConversion.String != "gauge" && result.RateConversion.String != "counters" {
		return result, fmt.Errorf("invalid rate conversion: %s, must be 'gauge' or 'counters'", result.RateConversion.String)
	}

	if result.TrendConversion.String != "gauges" && result.TrendConversion.String != "histogram" {
		return result, fmt.Errorf("invalid trend conversion: %s, must be 'gauges' or 'histogram'", result.TrendConversion.String)
	}

	dconf, _ := json.MarshalIndent(result, "", "  ")
	log.Debug("final_config", string(dconf))

	return result, nil
}

func envBool(env map[string]string, name string) (null.Bool, error) {
	if v, vDefined := env[name]; vDefined {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return null.NewBool(false, false), err
		}

		return null.BoolFrom(b), nil
	}
	return null.NewBool(false, false), nil
}

func parseEnvs(env map[string]string) (Config, error) {
	c := Config{
		Headers: make(map[string]string),
	}

	insecure, err := envBool(env, "K6_OTLP_INSECURE")
	if err == nil {
		c.Insecure = insecure
	}

	gzip, err := envBool(env, "K6_OTLP_GZIP")
	if err == nil {
		c.GZip = gzip
	}

	add, err := envBool(env, "K6_OTLP_ADD_ID_ATTRS")
	if err == nil {
		c.AddIDAttributes = add
	}

	if timeout, defined := env["K6_OTLP_TIMEOUT"]; defined {
		if err := c.Timeout.UnmarshalText([]byte(timeout)); err != nil {
			return c, err
		}
	}

	if pushInterval, pushIntervalDefined := env["K6_OTLP_PUSH_INTERVAL"]; pushIntervalDefined {
		if err := c.PushInterval.UnmarshalText([]byte(pushInterval)); err != nil {
			return c, err
		}
	}

	if url, urlDefined := env["K6_OTLP_SERVER_URL"]; urlDefined {
		c.ServerURL = null.StringFrom(url)
	}

	if convtype, defined := env["K6_OTLP_TREND_CONVERSION"]; defined {
		c.TrendConversion = null.StringFrom(convtype)
	}

	if convtype, defined := env["K6_OTLP_RATE_CONVERSION"]; defined {
		c.RateConversion = null.StringFrom(convtype)
	}

	if headers, headersDefined := env["K6_OTLP_HTTP_HEADERS"]; headersDefined {
		for _, kvPair := range strings.Split(headers, ",") {
			header := strings.Split(kvPair, ":")
			if len(header) != 2 {
				return c, fmt.Errorf("Provided header (%s) does not respect the expected format <key>:<value>", kvPair)
			}
			c.Headers[header[0]] = header[1]
		}
	}

	return c, nil
}

func parseJSON(data json.RawMessage) (Config, error) {
	var c Config
	err := json.Unmarshal(data, &c)
	return c, err
}

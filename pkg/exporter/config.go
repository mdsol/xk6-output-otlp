package exporter

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	Timeout         time.Duration
	Headers         http.Header
	GzipCompression bool
	Insecure        bool
	Endpoint        string
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("Endpoint should not be empty")
	}

	if err := c.validateURL(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateURL() error {
	_, err := url.Parse(c.Endpoint)
	if err != nil {
		return err
	}

	return nil
}

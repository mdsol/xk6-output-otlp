package extension

import (
	"github.com/mdsol/xk6-output-otlp/pkg/extension"

	"go.k6.io/k6/output"
)

func init() {
	output.RegisterExtension("otlp", func(p output.Params) (output.Output, error) {
		return extension.New(p)
	})
}

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/pflag"

	helmcli "helm.sh/helm/v3/pkg/cli"
)

type envSettings struct {
	*helmcli.EnvSettings
	dryRun bool
}

func newEnvSettings() *envSettings {
	envSettings := envSettings{}
	envSettings.EnvSettings = helmcli.New()
	return &envSettings
}

// AddBaseFlags binds base flags to the given flagset.
func (s *envSettings) AddBaseFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&s.dryRun, "dry-run", false, "simulate a command")
}

// AddFlags binds flags to the given flagset.
func (s *envSettings) AddFlags(fs *pflag.FlagSet) {
	s.AddBaseFlags(fs)
}

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Output(2, fmt.Sprintf(format, v...))
	}
}

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
)

var (
	resourceNames []string
)

type adoptOptions struct {
	dryRun        bool
	resourceNames []string
	args          []string
}

func newAdoptCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt [flags] RELEASE",
		Short: "path the api version of a resource",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("name of release and the chart has to be defined")
			}
			return nil
		},

		RunE: runAdopt,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	flags.StringArrayVar(&resourceNames, "names", []string{}, "the names of the recources to adopt")

	cmd.MarkFlagRequired("names")

	return cmd

}

func runAdopt(cmd *cobra.Command, args []string) error {

	adoptOptions := adoptOptions{
		dryRun:        settings.dryRun,
		resourceNames: resourceNames,
		args:          args,
	}
	return adopt(adoptOptions)
}

func adopt(opts adoptOptions) error {
	if opts.dryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	cfg := new(action.Configuration)
	if err := cfg.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"), debug); err != nil {
		return err
	}
	install := action.NewInstall(cfg)

	name, chartName, err := install.NameAndChart(opts.args)
	if err != nil {
		return err
	}
	install.ReleaseName = name

	cp, err := install.ChartPathOptions.LocateChart(chartName, settings.EnvSettings)
	if err != nil {
		return err
	}

	debug("CHART PATH: %s\n", cp)

	chrt, err := loader.Load(cp)
	if err != nil {
		return err
	}

	ts := cfg.Now()
	rel := &release.Release{
		Name:      name,
		Namespace: settings.Namespace(),
		Chart:     chrt,
		Config:    make(map[string]interface{}),
		Info: &release.Info{
			FirstDeployed: ts,
			LastDeployed:  ts,
			Status:        release.StatusUnknown,
		},
		Version: 1,
	}

	manifest, err := buildManifest(opts, cfg)
	if err != nil {
		return err
	}
	rel.Manifest = manifest
	rel.SetStatus(release.StatusDeployed, "Adoption complete")
	cfg.Releases.Create(rel)

	// TODO checks existing charts

	return nil
}

func buildManifest(opts adoptOptions, cfg *action.Configuration) (string, error) {
	b := bytes.NewBuffer(nil)

	for _, name := range opts.resourceNames {

		builder := resource.NewBuilder(cfg.RESTClientGetter)

		result := builder.
			Unstructured().
			NamespaceParam(settings.Namespace()).
			ExportParam(true).
			ResourceTypeOrNameArgs(true, name).
			Do()
		if result.Err() != nil {
			return "", result.Err()
		}
		object, err := result.Object()
		if err != nil {
			return "", err
		}
		us := object.(*unstructured.Unstructured)
		m, err := yaml.Marshal(us.Object)
		if err != nil {
			return "", err
		}

		content := string(m)

		if strings.TrimSpace(content) == "" {
			continue
		}
		fmt.Fprintf(b, "---\n# Source: %s\n%s\n", name, content)
	}

	return b.String(), nil
}

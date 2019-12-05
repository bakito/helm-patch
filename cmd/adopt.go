package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/bakito/helm-patch/pkg/types"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"
)

func newAdoptCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt [flags] [RELEASE] [CHART]",
		Short: "adopt existing resources into a chart",
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

	flags.StringArrayVarP(&names, "name", "n", []string{}, "the name(s) of the recources to adopt")
	flags.StringArrayVarP(&kinds, "kind", "k", []string{}, "the kind(s) of the recources to adopt")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("kind")

	return cmd

}

func runAdopt(cmd *cobra.Command, args []string) error {

	if len(names) != len(kinds) {
		return errors.New("The number of name args %d and kind args %d do not match")
	}

	opts := resourceNameOptions{
		Options: types.Options{
			DryRun:      settings.dryRun,
			ReleaseName: args[0],
		},
		names: names,
		kinds: kinds,
	}
	return adopt(opts)
}

func adopt(opts resourceNameOptions) error {
	if opts.DryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	cfg, err := settings.cfg()
	if err != nil {
		return err
	}
	install := action.NewInstall(cfg)

	name, chartName, err := install.NameAndChart([]string{opts.ReleaseName})
	if err != nil {
		return err
	}
	install.ReleaseName = name

	cp, err := install.ChartPathOptions.LocateChart(chartName, settings.EnvSettings)
	if err != nil {
		return err
	}

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

	manifest, resourceNames, err := buildManifest(opts, cfg)
	if err != nil {
		return err
	}

	list := action.NewList(cfg)
	results, err := list.Run()
	if err != nil {
		return err
	}

	for _, res := range results {
		manifests := releaseutil.SplitManifests(res.Manifest)

		for _, data := range manifests {
			resYaml := make(map[string]interface{})
			if err := yaml.Unmarshal([]byte(data), &resYaml); err != nil {
				return err
			}

			resource := types.ToResource(resYaml)
			if resource == nil {
				return nil
			}

			if _, ok := resourceNames[resource.KindName()]; ok {
				return fmt.Errorf("The resource '%s' is already contained within the chart: '%s-%s', name: '%s', version: %v",
					resource.KindName(), res.Chart.Name(), res.Chart.Metadata.Version, res.Name, res.Version)
			}
		}
	}

	if opts.DryRun {
		log.Printf("%s\n", manifest)
	} else {
		rel.Manifest = manifest
		rel.SetStatus(release.StatusDeployed, "Adoption complete")
		cfg.Releases.Create(rel)
	}

	return nil
}

func buildManifest(opts resourceNameOptions, cfg *action.Configuration) (string, map[string]bool, error) {
	b := bytes.NewBuffer(nil)

	resourceNames := make(map[string]bool)

	for i, name := range opts.names {
		resName := fmt.Sprintf("%s/%s", opts.kinds[i], name)

		builder := resource.NewBuilder(cfg.RESTClientGetter)

		result := builder.
			Unstructured().
			NamespaceParam(settings.Namespace()).
			ExportParam(true).
			ResourceTypeOrNameArgs(true, resName).
			Do()
		if result.Err() != nil {
			return "", resourceNames, result.Err()
		}
		object, err := result.Object()
		if err != nil {
			return "", resourceNames, err
		}
		us := object.(*unstructured.Unstructured)
		m, err := yaml.Marshal(us.Object)
		if err != nil {
			return "", resourceNames, err
		}

		content := string(m)

		if strings.TrimSpace(content) != "" {

			src := name

			if meta, ok2 := object.(metav1.Object); ok2 {
				src = object.GetObjectKind().GroupVersionKind().Kind + "/" + meta.GetName()
				resourceNames[object.GetObjectKind().GroupVersionKind().Kind+"/"+meta.GetName()] = true
			}

			fmt.Fprintf(b, "---\n# Exported form: %s\n%s\n", src, content)
		}
	}

	return b.String(), resourceNames, nil
}

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bakito/helm-patch/pkg/types"
	"github.com/bakito/helm-patch/pkg/util"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"
)

func newAdoptCmd() *cobra.Command {
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

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("kind")

	return cmd

}

func runAdopt(_ *cobra.Command, args []string) error {

	if len(names) != len(kinds) {
		return errors.New("the number of name args %d and kind args %d do not match")
	}

	opts := resourceNameOptions{
		Options: types.Options{
			DryRun:      settings.dryRun,
			ReleaseName: args[0],
		},
		names: names,
		kinds: kinds,
		chart: args[1],
	}
	return adopt(opts)
}

func adopt(opts resourceNameOptions) error {
	dr := ""
	if opts.DryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
		dr = "DRY-RUN "
	}

	cfg, err := settings.cfg()
	if err != nil {
		return err
	}
	install := action.NewInstall(cfg)

	name, chartName, err := install.NameAndChart([]string{opts.ReleaseName, opts.chart})
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

	for _, result := range results {
		manifests := util.SplitManifests(result.Manifest)

		for _, data := range manifests {
			resYaml := make(map[string]interface{})
			if err := yaml.Unmarshal([]byte(data), &resYaml); err != nil {
				return err
			}

			res := types.ToResource(resYaml)
			if res == nil {
				return nil
			}

			if _, ok := resourceNames[res.KindName()]; ok {
				return fmt.Errorf("%sThe resource '%s' is already contained within the chart: '%s-%s', name: '%s', version: %v",
					dr, res.KindName(), result.Chart.Name(), result.Chart.Metadata.Version, result.Name, result.Version)
			}
		}
	}

	if opts.DryRun {
		log.Printf("%s%s\n", dr, manifest)
	} else {
		rel.Manifest = manifest
		rel.SetStatus(release.StatusDeployed, "Adoption complete")
		_ = cfg.Releases.Create(rel)
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

			_, _ = fmt.Fprintf(b, "---\n# Exported form: %s\n%s\n", src, content)
		}
	}

	return b.String(), resourceNames, nil
}

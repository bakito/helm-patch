package cmd

import (
	"errors"
	"io"
	"log"
	"strings"

	"github.com/bakito/helm-patch/pkg/types"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	"sigs.k8s.io/yaml"
)

func newRmCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [flags] RELEASE",
		Short: "remove existing resources from a chart",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("name of release to be patched has to be defined")
			}
			return nil
		},

		RunE: runRm,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	flags.StringArrayVar(&resourceNames, "name", []string{}, "the name(s) of the recources to remove")

	cmd.MarkFlagRequired("name")

	return cmd

}

func runRm(cmd *cobra.Command, args []string) error {

	opts := resourceNameOptions{Options: types.Options{
		DryRun:      settings.dryRun,
		ReleaseName: args[0],
		Revision:    revision,
	},
		resourceNames: resourceNames,
		args:          args,
	}
	return remove(opts)
}

func remove(opts resourceNameOptions) error {
	if opts.DryRun {
		log.Println("NOTE: This is in dry-run mode, the following actions will not be executed.")
		log.Println("Run without --dry-run to take the actions described below:")
		log.Println()
	}

	cfg, err := settings.cfg()
	if err != nil {
		return err
	}

	releases, err := cfg.Releases.List(opts.Filter())

	if err != nil {
		return err
	}

	var rel *release.Release
	if len(releases) > 0 {
		rel = releases[len(releases)-1]
	}

	log.Printf("Processing release: '%s' with revision: %v\n", rel.Name, rel.Version)

	changed := false
	manifests := releaseutil.SplitManifests(rel.Manifest)

	for name, data := range manifests {
		resource := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(data), &resource); err != nil {
			return err
		}

		res := types.ToResource(resource)
		for _, rn := range opts.resourceNames {
			if strings.ToLower(rn) == strings.ToLower(res.KindName()) {
				delete(manifests, name)
				log.Printf("Remove resource '%s' from chart \n", res.KindName())
				changed = true
			}
		}

	}
	if changed {
		if !opts.DryRun {
			err = saveResource(manifests, rel, cfg)
			if err != nil {
				return err
			}
		}
		log.Printf("Release: '%s' with revision: %v patched successfully\n", rel.Name, rel.Version)

	} else {
		log.Print("Nothing to patch")
	}
	return nil
}

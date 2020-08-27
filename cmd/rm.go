package cmd

import (
	"errors"
	"log"
	"strings"

	"github.com/bakito/helm-patch/pkg/types"
	"github.com/bakito/helm-patch/pkg/util"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"
)

func newRmCmd() *cobra.Command {
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

	flags.StringArrayVarP(&names, "name", "n", []string{}, "the name(s) of the recources to remove")
	flags.StringArrayVarP(&kinds, "kind", "k", []string{}, "the kind(s) of the recources to remove")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("kind")

	return cmd

}

func runRm(_ *cobra.Command, args []string) error {
	if len(names) != len(kinds) {
		return errors.New("the number of name args %d and kind args %d do not match")
	}

	opts := resourceNameOptions{Options: types.Options{
		DryRun:      settings.dryRun,
		ReleaseName: args[0],
		Revision:    revision,
	},
		names: names,
		kinds: kinds,
	}
	return remove(opts)
}

func remove(opts resourceNameOptions) error {
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

	releases, err := cfg.Releases.List(opts.Filter())

	if err != nil {
		return err
	}

	var rel *release.Release
	if len(releases) > 0 {
		rel = releases[len(releases)-1]
	} else {
		log.Printf("No release found\n")
		return nil
	}

	log.Printf("%sProcessing release: '%s' with revision: %v\n", dr, rel.Name, rel.Version)

	changed := false
	manifests := util.SplitManifests(rel.Manifest)

	for manifestNamew, data := range manifests {
		resource := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(data), &resource); err != nil {
			return err
		}

		res := types.ToResource(resource)
		for i, name := range opts.names {
			if strings.ToLower(name) == strings.ToLower(res.Name()) && strings.ToLower(opts.kinds[i]) == strings.ToLower(res.Kind()) {
				delete(manifests, manifestNamew)
				log.Printf("%sRemove resource '%s' from chart \n", dr, res.KindName())
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
		log.Printf("%sRelease: '%s' with revision: %v patched successfully\n", dr, rel.Name, rel.Version)

	} else {
		log.Printf("%sResources not found\n", dr)
	}
	return nil
}

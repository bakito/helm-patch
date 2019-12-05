package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/bakito/helm-patch/pkg/types"
	"github.com/bakito/helm-patch/pkg/util"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"
)

func newAPICmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api [flags] RELEASE",
		Short: "path the api version of a resource",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("name of release to be patched has to be defined")
			}
			return nil
		},

		RunE: runAPI,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	flags.StringVarP(&kind, "kind", "k", "", "the kind to patch the api version")
	flags.StringVar(&from, "from", "", "the api version that has to be replaced")
	flags.StringVar(&to, "to", "", "the api version to be set")
	flags.StringVarP(&name, "name", "n", "", "the name of the resource")
	flags.IntVar(&revision, "revision", -1, "the revision of the release to path")

	cmd.MarkFlagRequired("kind")
	cmd.MarkFlagRequired("to")

	return cmd

}

func runAPI(cmd *cobra.Command, args []string) error {

	apiOptions := apiOptions{
		Options: types.Options{
			DryRun:      settings.dryRun,
			ReleaseName: args[0],
			Revision:    revision,
		},
		kind:         kind,
		from:         from,
		to:           to,
		resourceName: name,
	}
	return patchAPI(apiOptions)
}

func patchAPI(opts apiOptions) error {
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
	}

	log.Printf("%sProcessing release: '%s' with revision: %v\n", dr, rel.Name, rel.Version)

	changed := false
	manifests := util.SplitManifests(rel.Manifest)

	for name, data := range manifests {
		resource := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(data), &resource); err != nil {
			return err
		}

		if i := apiInfo(opts, resource); i != nil {
			p, err := patchManifest(opts, resource, i)
			if err != nil {
				return err
			} else if p != "" {
				manifests[name] = p
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
		log.Printf("%sNothing to patch\n", dr)
	}
	return nil
}

func saveResource(manifests map[string]string, rel *release.Release, cfg *action.Configuration) error {
	b := bytes.NewBuffer(nil)
	for name, content := range manifests {
		if strings.TrimSpace(content) == "" {
			continue
		}
		fmt.Fprintf(b, "---\n# Source: %s\n%s\n", name, content)
	}
	rel.Manifest = b.String()
	return cfg.Releases.Update(rel)
}

func patchManifest(opts apiOptions, resource map[string]interface{}, r types.Resource) (string, error) {
	log.Printf("Patching kind: %s name: %s from apiVersion: %s to apiVersion: %s\n", r.Kind(), r.Name(), opts.from, opts.to)
	resource["apiVersion"] = opts.to

	m, err := yaml.Marshal(resource)
	if err == nil {
		return string(m), nil
	}
	return "", nil
}

func apiInfo(opts apiOptions, yaml map[string]interface{}) types.Resource {
	resource := types.ToResource(yaml)
	if resource == nil {
		return nil
	}

	name := resource.Name()
	if name == "" || (name != opts.resourceName && opts.resourceName != "") {
		return nil
	}

	k := resource.Kind()
	if k == "" || k != opts.kind {
		return nil
	}

	version := resource.GroupVersion()
	if version == "" || (version != opts.from && opts.from != "") {
		return nil
	}

	return resource
}

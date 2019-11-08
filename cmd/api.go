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
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
)

var (
	kind     string
	from     string
	to       string
	name     string
	revision int
)

type ApiOptions struct {
	DryRun       bool
	Kind         string
	From         string
	To           string
	ResourceName string
	ReleaseName  string
	Revision     int
}

type resourceInfo struct {
	apiVersion string
	kind       string
	name       string
}

func newApiCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api [flags] RELEASE",
		Short: "path the api version of a resource",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("name of release to be patched has to be defined")
			}
			return nil
		},

		RunE: runApi,
	}

	flags := cmd.Flags()
	settings.AddFlags(flags)

	flags.StringVar(&kind, "kind", "", "the kind to patch the api version")
	flags.StringVar(&from, "from", "", "the api version that has to be replaced")
	flags.StringVar(&to, "to", "", "the api version to be set")
	flags.StringVar(&name, "name", "", "the name of the resource")
	flags.IntVar(&revision, "revision", -1, "the revision of the release to path")

	cmd.MarkFlagRequired("kind")
	cmd.MarkFlagRequired("to")

	return cmd

}

func runApi(cmd *cobra.Command, args []string) error {

	apiOptions := ApiOptions{
		DryRun:       settings.dryRun,
		Kind:         kind,
		From:         from,
		To:           to,
		ResourceName: name,
		ReleaseName:  args[0],
		Revision:     revision,
	}
	return PatchApi(apiOptions)
}

func PatchApi(opts ApiOptions) error {
	if opts.DryRun {
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

	releases, err := cfg.Releases.List(opts.filter)
	if err != nil {
		return err
	}

	var rel *release.Release
	if len(releases) > 0 {
		rel = releases[len(releases)-1]
	}

	log.Printf("Processing release: '%s' with revision: %v\n", rel.Name, rel.Version)

	for _, r := range releases {
		changed := false
		manifests := releaseutil.SplitManifests(r.Manifest)
		for name, data := range manifests {
			resource := make(map[string]interface{})
			if err := yaml.Unmarshal([]byte(data), &resource); err != nil {
				return err
			}

			if i := info(opts, resource); i != nil {
				resource["apiVersion"] = opts.To
				log.Printf("Patching kind: %s name: %s from apiVersion: %s to apiVersion: %s\n", i.kind, i.name, i.apiVersion, opts.To)
				changed = true
				if !opts.DryRun {
					m, err := yaml.Marshal(resource)
					if err != nil {
						return err
					}
					manifests[name] = string(m)
				}
			}
		}

		if changed {
			if !opts.DryRun {
				b := bytes.NewBuffer(nil)
				for name, content := range manifests {
					if strings.TrimSpace(content) == "" {
						continue
					}
					fmt.Fprintf(b, "---\n# Source: %s\n%s\n", name, content)
				}
				r.Manifest = b.String()
				err = cfg.Releases.Update(r)
				if err != nil {
					return err
				}
			}
			log.Printf("Release: '%s' with revision: %v patched successfully\n", rel.Name, rel.Version)

		} else {
			log.Print("Nothing to patch")
		}
	}
	return nil
}

func info(opts ApiOptions, resource map[string]interface{}) *resourceInfo {
	k, ok := resource["kind"]
	if !ok || k != opts.Kind {
		return nil
	}

	version, ok := resource["apiVersion"]
	if !ok || (version == opts.From && opts.From != "") {
		return nil
	}

	metadata, ok := resource["metadata"]
	if !ok {
		return nil
	}

	name, ok := metadata.(map[interface{}]interface{})["name"]
	if !ok || (name != opts.ResourceName && opts.ResourceName != "") {
		return nil
	}

	return &resourceInfo{
		kind:       k.(string),
		apiVersion: version.(string),
		name:       name.(string),
	}
}

func (opts *ApiOptions) filter(r *release.Release) bool {
	if r == nil || r.Name != opts.ReleaseName {
		return false
	}

	if opts.Revision > 0 {
		return r.Version == opts.Revision
	}
	return true
}

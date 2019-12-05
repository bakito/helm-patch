package cmd

import "github.com/bakito/helm-patch/pkg/types"

var (
	revision int

	name  string
	names []string
	kind  string
	kinds []string

	from string
	to   string
)

type resourceNameOptions struct {
	types.Options
	names []string
	kinds []string
}

type apiOptions struct {
	types.Options
	kind         string
	from         string
	to           string
	resourceName string
}

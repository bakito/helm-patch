package cmd

import (
	"testing"

	. "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
	"helm.sh/helm/v3/pkg/release"
)

var infoDataset = []struct {
	opts     apiOptions
	resource map[string]interface{}
	expected *resourceInfo
}{
	{
		apiOptions{kind: "Bar"},
		map[string]interface{}{"kind": "Foo"},
		nil,
	},
	{
		apiOptions{kind: "Foo"},
		map[string]interface{}{"kind": "Foo"},
		nil,
	},
	{
		apiOptions{kind: "Foo", from: ""},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1"},
		nil,
	},
	{
		apiOptions{kind: "Foo", from: "v1"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1"},
		nil,
	},
	{
		apiOptions{kind: "Foo", from: "v2"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1"},
		nil,
	},
	{
		apiOptions{kind: "Foo", from: "v1"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1", "metadata": map[string]interface{}{}},
		nil,
	},
	{
		apiOptions{kind: "Foo", from: "v1"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1", "metadata": map[string]interface{}{"name": "abc"}},
		&resourceInfo{apiVersion: "v1", kind: "Foo", name: "abc"},
	},
	{
		apiOptions{kind: "Foo", from: "v1", resourceName: "xyz"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1", "metadata": map[string]interface{}{"name": "abc"}},
		nil,
	},
}

func Test_info(t *testing.T) {
	for i, ds := range infoDataset {

		ri := info(ds.opts, ds.resource)
		if ds.expected == nil {
			Assert(t, is.Nil(ri), "InfoDataset #%v: %v", i, ds)
		} else {
			Assert(t, ri != nil, "InfoDataset #%v: %v", i, ds)
			Assert(t, is.Equal(ds.expected.apiVersion, ri.GroupVersion()), "InfoDataset #%v: %v", i, ds)
			Assert(t, is.Equal(ds.expected.kind, ri.Kind()), "InfoDataset #%v: %v", i, ds)
			Assert(t, is.Equal(ds.expected.name, ri.Name()), "InfoDataset #%v: %v", i, ds)
		}
	}
}

var filterDataset = []struct {
	opts     apiOptions
	release  *release.Release
	expected bool
}{
	{
		apiOptions{},
		&release.Release{},
		false,
	}, {
		apiOptions{releaseName: "abc"},
		&release.Release{Name: "abc"},
		true,
	},
	{
		apiOptions{releaseName: "abc"},
		&release.Release{Name: "xyz"},
		false,
	},
	{
		apiOptions{releaseName: "abc", revision: 1},
		&release.Release{Name: "abc", Version: 1},
		true,
	},
	{
		apiOptions{releaseName: "abc", revision: 1},
		&release.Release{Name: "abc", Version: 2},
		false,
	},
}

func Test_filter(t *testing.T) {
	for i, ds := range filterDataset {
		match := ds.opts.filter(ds.release)
		Assert(t, is.Equal(ds.expected, match), "FilterDataset #%v: %v", i, ds)
	}
}

type resourceInfo struct {
	apiVersion string
	kind       string
	name       string
}

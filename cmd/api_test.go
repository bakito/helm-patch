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
		apiOptions{Kind: "Bar"},
		map[string]interface{}{"kind": "Foo"},
		nil,
	},
	{
		apiOptions{Kind: "Foo"},
		map[string]interface{}{"kind": "Foo"},
		nil,
	},
	{
		apiOptions{Kind: "Foo", From: ""},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1"},
		nil,
	},
	{
		apiOptions{Kind: "Foo", From: "v1"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1"},
		nil,
	},
	{
		apiOptions{Kind: "Foo", From: "v2"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1"},
		nil,
	},
	{
		apiOptions{Kind: "Foo", From: "v1"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1", "metadata": map[interface{}]interface{}{}},
		nil,
	},
	{
		apiOptions{Kind: "Foo", From: "v1"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1", "metadata": map[interface{}]interface{}{"name": "abc"}},
		&resourceInfo{apiVersion: "v1", kind: "Foo", name: "abc"},
	},
	{
		apiOptions{Kind: "Foo", From: "v1", ResourceName: "xyz"},
		map[string]interface{}{"kind": "Foo", "apiVersion": "v1", "metadata": map[interface{}]interface{}{"name": "abc"}},
		nil,
	},
}

func Test_info(t *testing.T) {
	for i, ds := range infoDataset {
		ri := info(ds.opts, ds.resource)
		if ds.expected == nil {
			Assert(t, is.Nil(ds.expected), "InfoDataset #%v: %v", i, ds)
		} else {
			Assert(t, is.Equal(ds.expected.apiVersion, ri.apiVersion), "InfoDataset #%v: %v", i, ds)
			Assert(t, is.Equal(ds.expected.kind, ri.kind), "InfoDataset #%v: %v", i, ds)
			Assert(t, is.Equal(ds.expected.name, ri.name), "InfoDataset #%v: %v", i, ds)
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
		apiOptions{ReleaseName: "abc"},
		&release.Release{Name: "abc"},
		true,
	},
	{
		apiOptions{ReleaseName: "abc"},
		&release.Release{Name: "xyz"},
		false,
	},
	{
		apiOptions{ReleaseName: "abc", Revision: 1},
		&release.Release{Name: "abc", Version: 1},
		true,
	},
	{
		apiOptions{ReleaseName: "abc", Revision: 1},
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

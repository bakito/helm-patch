package cmd

import (
	"testing"

	. "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
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

func Test_apiInfo(t *testing.T) {
	for i, ds := range infoDataset {

		ri := apiInfo(ds.opts, ds.resource)
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

type resourceInfo struct {
	apiVersion string
	kind       string
	name       string
}

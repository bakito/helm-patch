package types

import (
	"testing"

	. "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
	"helm.sh/helm/v3/pkg/release"
)

var filterDataset = []struct {
	opts     Options
	release  *release.Release
	expected bool
}{
	{
		Options{},
		&release.Release{},
		false,
	}, {
		Options{ReleaseName: "abc"},
		&release.Release{Name: "abc"},
		true,
	},
	{
		Options{ReleaseName: "abc"},
		&release.Release{Name: "xyz"},
		false,
	},
	{
		Options{ReleaseName: "abc", Revision: 1},
		&release.Release{Name: "abc", Version: 1},
		true,
	},
	{
		Options{ReleaseName: "abc", Revision: 1},
		&release.Release{Name: "abc", Version: 2},
		false,
	},
}

func Test_filter(t *testing.T) {
	for i, ds := range filterDataset {
		f := ds.opts.Filter()
		match := f(ds.release)
		Assert(t, is.Equal(ds.expected, match), "FilterDataset #%v: %v", i, ds)
	}
}

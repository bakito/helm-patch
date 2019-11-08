package cmd

import (
	"testing"

	. "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

var tests = []struct {
	opts     ApiOptions
	resource map[string]interface{}
	expected *resourceInfo
}{
	{ApiOptions{Kind: "Bar"}, map[string]interface{}{"kind": "Foo"}, nil},
	{ApiOptions{Kind: "Foo"}, map[string]interface{}{"kind": "Foo"}, nil},
}

func Test_info(t *testing.T) {
	for _, tt := range tests {
		i := info(tt.opts, tt.resource)

		Assert(t, is.Equal(tt.expected, i))
	}
}

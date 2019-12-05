package util

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	sourcePrefix = "# Source: "
)

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

// SplitManifests takes a string of manifest and returns a map contains individual manifests
func SplitManifests(bigFile string) map[string]string {
	// Basically, we're quickly splitting a stream of YAML documents into an
	// array of YAML docs.
	tpl := "manifest-%d"
	res := map[string]string{}
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	var count int
	for _, d := range docs {
		if d == "" {
			continue
		}

		d = strings.TrimSpace(d)
		lines := strings.Split(d, "\n")
		var key string
		if strings.HasPrefix(lines[0], sourcePrefix) {
			key = strings.Replace(lines[0], sourcePrefix, "", 1)
			d = strings.Join(lines[1:], "\n")
		} else {
			key = fmt.Sprintf(tpl, count)
		}

		res[key] = d
		count = count + 1
	}
	return res
}

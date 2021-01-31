package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cluttercode/clutter/internal/pkg/scanner"
)

func TestParseElement(t *testing.T) {
	tests := []struct {
		text   string
		path   string
		err    bool
		name   string
		attrs  map[string]string
		search bool
	}{
		{
			err: true,
		},
		{
			text: "meow",
			name: "meow",
		},
		{
			text: ".meow",
			name: "meow",
			attrs: map[string]string{
				"scope": "dir/file",
			},
		},
		{
			text: "./meow",
			name: "meow",
			attrs: map[string]string{
				"scope": "dir/",
			},
		},
		{
			text: "@attr meow",
			name: "meow",
			attrs: map[string]string{
				"attr": "",
			},
		},
		{
			text: "@attr ./meow",
			name: "meow",
			attrs: map[string]string{
				"attr":  "",
				"scope": "dir/",
			},
		},
		{
			text: "@attr1 meow attr2",
			name: "meow",
			attrs: map[string]string{
				"attr1": "",
				"attr2": "",
			},
		},
		{
			text: "@who=midnight meow",
			name: "meow",
			attrs: map[string]string{
				"who": "midnight",
			},
		},
		{
			text: "@who=midnight @when=now meow",
			name: "meow",
			attrs: map[string]string{
				"who":  "midnight",
				"when": "now",
			},
		},
		{
			text: "@who=midnight @when=\"now and then\" meow",
			name: "meow",
			attrs: map[string]string{
				"who":  "midnight",
				"when": "now and then",
			},
		},
		{
			text: "@who=midnight @when=\"now and then\" meow",
			name: "meow",
			attrs: map[string]string{
				"who":  "midnight",
				"when": "now and then",
			},
		},
		{
			text: "@who=midnight meow when=now",
			name: "meow",
			attrs: map[string]string{
				"who":  "midnight",
				"when": "now",
			},
		},
		{
			text: "@who=midnight .meow when=now",
			name: "meow",
			attrs: map[string]string{
				"who":   "midnight",
				"when":  "now",
				"scope": "dir/file",
			},
		},
		{
			text: "meow who=midnight when=now",
			name: "meow",
			attrs: map[string]string{
				"who":  "midnight",
				"when": "now",
			},
		},
		{
			text: "@who=zumi meow who=midnight when=now",
			err:  true,
		},
		{
			text: "@scope=x .meow",
			err:  true,
		},
		{
			text: "meow who=midnight scope=.",
			name: "meow",
			attrs: map[string]string{
				"who":   "midnight",
				"scope": "dir/file",
			},
		},
		{
			text: "meow who=midnight scope=\"./\"",
			name: "meow",
			attrs: map[string]string{
				"who":   "midnight",
				"scope": "dir/",
			},
		},
		{
			text: "./meow",
			name: "meow",
			path: "/",
			attrs: map[string]string{
				"scope": "/",
			},
		},
		{
			text: "./meow",
			name: "meow",
			path: "some_file_at_root",
		},
		{
			text: "? _ x=\"any*\"",
			name: "_",
			attrs: map[string]string{
				"search": "exact",
				"x":      "any*",
			},
			search: true,
		},
		{
			text: "?g bla x=\"any*\"",
			name: "bla",
			attrs: map[string]string{
				"search": "glob",
				"x":      "any*",
			},
			search: true,
		},
		{
			text: "?g \"*\" x=\"any*\"",
			name: "*",
			attrs: map[string]string{
				"search": "glob",
				"x":      "any*",
			},
			search: true,
		},
		{
			text: "?g \"*\"",
			name: "*",
			attrs: map[string]string{
				"search": "glob",
			},
			search: true,
		},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			path := test.path
			if path == "" {
				path = "dir/file"
			}

			ent, err := ParseElement(
				&scanner.RawElement{
					Text: test.text,
					Loc:  scanner.Loc{Path: path},
				},
			)

			if test.err {
				assert.Error(t, err)
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, test.name, ent.Name)
			assert.EqualValues(t, test.attrs, ent.Attrs)

			_, s := ent.IsSearch()
			assert.Equal(t, test.search, s)
		})
	}
}

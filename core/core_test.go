package core_test

import (
	"testing"

	"github.com/smartystreets/assertions"

	"github.com/drgrib/alfred-bear/core"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		expected core.Query
	}{
		{
			name:     "empty arg",
			arg:      "",
			expected: core.Query{Tokens: []string{""}},
		},
		{
			name: "single word",
			arg:  "hello",
			expected: core.Query{
				Tokens:     []string{"hello"},
				LastToken:  "hello",
				WordString: "hello",
			},
		},
		{
			name: "two words",
			arg:  "hello world",
			expected: core.Query{
				Tokens:     []string{"hello", "world"},
				LastToken:  "world",
				WordString: "hello world",
			},
		},
		{
			name: "multiple words with bad spacing",
			arg:  "hello  \t world",
			expected: core.Query{
				Tokens:     []string{"hello", "world"},
				LastToken:  "world",
				WordString: "hello world",
			},
		},
		{
			name: "argument contains tag",
			arg:  "hello #inbox stuff",
			expected: core.Query{
				Tokens:     []string{"hello", "#inbox", "stuff"},
				Tags:       []string{"#inbox"},
				LastToken:  "stuff",
				WordString: "hello stuff",
			},
		},
		{
			name: "tag is the last token",
			arg:  "hello #inbox",
			expected: core.Query{
				Tokens:     []string{"hello", "#inbox"},
				Tags:       []string{"#inbox"},
				LastToken:  "#inbox",
				WordString: "hello",
			},
		},
		{
			name: "multiword tag",
			arg:  "oh boy #hello tag#",
			expected: core.Query{
				Tokens:     []string{"oh", "boy", "#hello", "tag#"},
				Tags:       []string{"#hello tag#"},
				LastToken:  "tag#",
				WordString: "oh boy",
			},
		},
		{
			name: "multiword tag with later text",
			arg:  "oh boy #hello tag# more text",
			expected: core.Query{
				Tokens:     []string{"oh", "boy", "#hello", "tag#", "more", "text"},
				Tags:       []string{"#hello tag#"},
				LastToken:  "text",
				WordString: "oh boy more text",
			},
		},
	}

	for _, test := range tests {
		// nolint: scopelint
		t.Run(test.name, func(t *testing.T) {
			ok, msg := assertions.So(
				core.ParseQuery(test.arg),
				assertions.ShouldResemble,
				test.expected,
			)
			if !ok {
				t.Error(msg)
			}
		})
	}
}

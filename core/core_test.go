package core_test

import (
	"testing"

	"github.com/drgrib/alfred-bear/core"
	. "github.com/smartystreets/assertions"
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
			expected: core.Query{},
		},
		{
			name: "single word",
			arg:  "hello",
			expected: core.Query{
				Tokens:     []string{"hello"},
				Tags:       nil,
				LastToken:  "hello",
				WordString: "hello",
			},
		},
		{
			name: "two words",
			arg:  "hello world",
			expected: core.Query{
				Tokens:     []string{"hello", "world"},
				Tags:       nil,
				LastToken:  "world",
				WordString: "hello world",
			},
		},
		{
			name: "two words with bad spacing",
			arg:  "hello  \t world",
			expected: core.Query{
				Tokens:     []string{"hello", "world"},
				Tags:       nil,
				LastToken:  "world",
				WordString: "hello world",
			},
		},
		{
			name: "argument contains tag",
			arg:  "hello #inbox stuff",
			expected: core.Query{
				Tokens:     []string{"hello", "#inbox", "stuff"},
				Tags:       []string{"inbox"},
				LastToken:  "stuff",
				WordString: "hello #inbox stuff",
			},
		},
		{
			name: "tag is the last token",
			arg:  "hello #inbox",
			expected: core.Query{
				Tokens:     []string{"hello", "#inbox"},
				Tags:       []string{"inbox"},
				LastToken:  "#inbox",
				WordString: "hello #inbox",
			},
		},
	}

	for _, test := range tests {
		// nolint: scopelint
		t.Run(test.name, func(t *testing.T) {
			So(core.ParseQuery(test.arg), ShouldResemble, test.expected)
		})
	}
}

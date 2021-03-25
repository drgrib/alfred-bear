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
			expected: core.Query{Tokens: []string{""}, Tags: []string{}},
		},
		{
			name: "single word",
			arg:  "hello",
			expected: core.Query{
				Tokens:     []string{"hello"},
				Tags:       []string{},
				LastToken:  "hello",
				WordString: "hello",
			},
		},
		{
			name: "two words",
			arg:  "hello world",
			expected: core.Query{
				Tokens:     []string{"hello", "world"},
				Tags:       []string{},
				LastToken:  "world",
				WordString: "hello world",
			},
		},
		{
			name: "multiple words with bad spacing",
			arg:  "hello  \t world",
			expected: core.Query{
				Tokens:     []string{"hello", "world"},
				Tags:       []string{},
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
				WordString: "hello stuff",
			},
		},
		{
			name: "tag is the last token",
			arg:  "hello #inbox",
			expected: core.Query{
				Tokens:     []string{"hello", "#inbox"},
				Tags:       []string{"inbox"},
				LastToken:  "#inbox",
				WordString: "hello",
			},
		},
		{
			name: "multiword tag",
			arg:  "#hello tag#",
			expected: core.Query{
				Tokens:     []string{"#hello", "tag#"},
				Tags:       []string{"hello tag"},
				LastToken:  "tag#",
				WordString: "",
			},
		},
		{
			name: "complex multiword tag",
			arg:  "hi #test tag# #test2 there",
			expected: core.Query{
				Tokens:     []string{"hi", "#test", "tag#", "#test2", "there"},
				Tags:       []string{"test tag", "test2"},
				LastToken:  "there",
				WordString: "hi there",
			},
		},
	}

	for _, test := range tests {
		// nolint: scopelint
		t.Run(test.name, func(t *testing.T) {
			if ok, msg := So(core.ParseQuery(test.arg), ShouldResemble, test.expected); !ok {
				t.Error(msg)
			}
		})
	}
}

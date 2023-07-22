package db_test

import (
	"testing"

	"github.com/smartystreets/assertions"

	"github.com/drgrib/alfred-bear/db"
)

func TestRemoveTagHashes(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{
			tag:      "#simple",
			expected: "simple",
		},
		{
			tag:      "#multi word#",
			expected: "multi word",
		},
		{
			tag:      "#multi word long#",
			expected: "multi word long",
		},
	}

	for _, test := range tests {
		// nolint: scopelint
		t.Run(test.tag, func(t *testing.T) {
			ok, msg := assertions.So(
				db.RemoveTagHashes(test.tag),
				assertions.ShouldResemble,
				test.expected,
			)
			if !ok {
				t.Error(msg)
			}
		})
	}
}

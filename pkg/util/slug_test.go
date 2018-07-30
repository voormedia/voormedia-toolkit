package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	in  string
	out string
}

var testCases = []testCase{
	{"hi there", "hi-there"},
	{"Hello World!", "hello-world"},
	{"'What a wonderful day...'", "what-a-wonderful-day"},
}

func TestSlugify(t *testing.T) {
	for _, tc := range testCases {
		assert.Equal(t, tc.out, Slugify(tc.in))
	}
}

package util

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	buf := bytes.NewBufferString("")
	log := NewLogger("foo")
	log.out = buf

	log.Log("hi", "there")
	log.Note("hello", "there")
	log.Success("hi", "there")
	log.Warn("hello", "there")
	log.Error("hi", "there")

	assert.Equal(t, strings.Join([]string{
		"foo: hi there\n",
		"foo: hello there\n",
		"foo: hi there\n",
		"foo: hello there\n",
		"foo: hi there\n",
	}, ""), buf.String())
}

package api

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestFilter(t *testing.T) {
	m := map[string]string{
		"a": "1234567890987654321#",
		"b": "1234567890987654321",
		"c": "1234567890987654321##",
	}
	t.Log(Filter(m))
	assert.Equal(t, m["a"], "1234567890987654321…")
	assert.Equal(t, m["b"], "1234567890987654321")
	assert.Equal(t, m["c"], "1234567890987654321…")
}

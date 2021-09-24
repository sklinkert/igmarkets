package igmarkets

import (
	"github.com/AMekss/assert"
	"testing"
)

func TestNew(t *testing.T) {
	igm := New(DemoAPIURL, "", "", "", "")
	assert.False(t, igm == nil)
}

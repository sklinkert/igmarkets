package igmarkets

import (
	"github.com/AMekss/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	igm := New(DemoAPIURL, "", "", "", "", time.Second)
	assert.False(t, igm == nil)
}

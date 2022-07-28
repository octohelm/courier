package core

import (
	"testing"
)

func TestTransformerCache(t *testing.T) {
	for name, tf := range mgrDefault.transformerSet {
		t.Log(name, tf)
	}
}

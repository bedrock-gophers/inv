package inv

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestDropperRegisteredBeforeRegistryFinalization(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	name, properties := (dropper{}).EncodeBlock()
	block, ok := world.BlockByName(name, properties)
	if !ok {
		t.Fatal("dropper block was not registered")
	}
	if _, ok := block.(dropper); !ok {
		t.Fatalf("dropper block has type %T", block)
	}
}

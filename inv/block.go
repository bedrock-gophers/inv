package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
)

func init() {
	// Dragonfly 0.11 finalizes the block registry on server creation and forbids lookups before then. Register the
	// dropper implementation directly while the default registry is still in its mutable package-init phase.
	world.RegisterBlock(dropper{})
}

type nopContainer struct{}

func (nopContainer) AddViewer(block.ContainerViewer, *world.World, cube.Pos)    {}
func (nopContainer) RemoveViewer(block.ContainerViewer, *world.World, cube.Pos) {}
func (nopContainer) Inventory() *inventory.Inventory {
	return inventory.New(69, func(slot int, before, after item.Stack) {})
}

type dropper struct{ nopContainer }

func (d dropper) Hash() (uint64, uint64) { return 932472, 932473 }
func (dropper) Model() world.BlockModel  { return model.Solid{} }
func (dropper) EncodeBlock() (string, map[string]any) {
	return "minecraft:dropper", map[string]any{"facing_direction": int32(0), "toggle_bit": false}
}

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
	for _, b := range []world.Block{hopper{}, dropper{}} {
		if bl, ok := world.BlockByName(b.EncodeBlock()); ok {
			if bl.Hash() != b.Hash() {
				world.RegisterBlock(b)
			}
		}
	}
}

type nopContainer struct{}

func (nopContainer) AddViewer(block.ContainerViewer, *world.World, cube.Pos)    {}
func (nopContainer) RemoveViewer(block.ContainerViewer, *world.World, cube.Pos) {}
func (nopContainer) Inventory() *inventory.Inventory {
	return inventory.New(69, func(slot int, before, after item.Stack) {})
}

type hopper struct{ nopContainer }
type dropper struct{ nopContainer }

func (h hopper) Hash() uint64          { return blockHash(h, 932473) }
func (hopper) Model() world.BlockModel { return model.Solid{} }
func (hopper) EncodeBlock() (string, map[string]any) {
	return "minecraft:hopper", map[string]any{"facing_direction": int32(0), "toggle_bit": false}
}

func (d dropper) Hash() uint64          { return blockHash(d, 932472) }
func (dropper) Model() world.BlockModel { return model.Solid{} }
func (dropper) EncodeBlock() (string, map[string]any) {
	return "minecraft:dropper", map[string]any{"facing_direction": int32(0), "toggle_bit": false}
}

func blockHash(b world.Block, base uint64) uint64 {
	if _, ok := world.BlockByName(b.EncodeBlock()); !ok {
		return 0
	}
	return base
}

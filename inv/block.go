package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"math"
)

func init() {
	for _, b := range []world.Block{dropper{}} {
		if bl, ok := world.BlockByName(b.EncodeBlock()); ok {
			hash, _ := bl.Hash()
			if hash == math.MaxUint64 {
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

type dropper struct{ nopContainer }

func (d dropper) Hash() (uint64, uint64) { return 932472, 932473 }
func (dropper) Model() world.BlockModel  { return model.Solid{} }
func (dropper) EncodeBlock() (string, map[string]any) {
	return "minecraft:dropper", map[string]any{"facing_direction": int32(0), "toggle_bit": false}
}

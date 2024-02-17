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
		if _, ok := world.BlockByName(b.EncodeBlock()); !ok {
			world.RegisterBlock(b)
		}
	}
}

type nopContainer struct{}

func (nopContainer) AddViewer(v block.ContainerViewer, w *world.World, pos cube.Pos)    {}
func (nopContainer) RemoveViewer(v block.ContainerViewer, w *world.World, pos cube.Pos) {}
func (nopContainer) Inventory() *inventory.Inventory {
	return inventory.New(69, func(slot int, before, after item.Stack) {})
}

type hopper struct{ nopContainer }
type dropper struct{ nopContainer }

func (hopper) Hash() uint64            { return 913214 }
func (hopper) Model() world.BlockModel { return model.Solid{} }
func (hopper) EncodeBlock() (string, map[string]any) {
	return "minecraft:hopper", map[string]any{"facing_direction": int32(0), "toggle_bit": false}
}

func (dropper) Hash() uint64            { return 913215 }
func (dropper) Model() world.BlockModel { return model.Solid{} }
func (dropper) EncodeBlock() (string, map[string]any) {
	return "minecraft:dropper", map[string]any{"facing_direction": int32(0), "toggle_bit": false}
}

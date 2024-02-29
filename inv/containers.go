package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Container interface {
	Block() world.Block
	Type() int
	Size() int
}

type ChestContainer struct{ DoubleChest bool }

func (ChestContainer) Block() world.Block {
	b := block.NewChest()
	b.Facing = 1
	return b
}
func (ChestContainer) Type() int { return protocol.ContainerTypeContainer }

func (c ChestContainer) Size() int {
	if c.DoubleChest {
		return 54
	}
	return 27
}

type HopperContainer struct{}

func (HopperContainer) Block() world.Block { return hopper{} }
func (HopperContainer) Type() int          { return protocol.ContainerTypeHopper }
func (HopperContainer) Size() int          { return 5 }

type DropperContainer struct{}

func (DropperContainer) Block() world.Block { return dropper{} }
func (DropperContainer) Type() int          { return protocol.ContainerTypeDropper }
func (DropperContainer) Size() int          { return 9 }

type BarrelContainer struct{}

func (BarrelContainer) Block() world.Block { return block.NewBarrel() }
func (BarrelContainer) Type() int          { return protocol.ContainerBarrel }
func (BarrelContainer) Size() int          { return 27 }

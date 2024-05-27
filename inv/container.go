package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// containerPos is the main container position. It is used as bait, for dragonfly to think that the player is
// using that container.
var containerPos cube.Pos

// Container represents a container that can be opened by a player. Containers are blocks that can store items
type Container interface {
	Block() world.Block
	Type() int
	Size() int
}

// ContainerChest represents a chest container. It can be a single chest or a double chest.
type ContainerChest struct{ DoubleChest bool }

// Block ...
func (c ContainerChest) Block() world.Block {
	b := block.NewChest()
	b.Facing = 1
	return b
}

// Type ...
func (ContainerChest) Type() int { return protocol.ContainerTypeContainer }

// Size ...
func (c ContainerChest) Size() int {
	if c.DoubleChest {
		return 54
	}
	return 27
}

// ContainerHopper represents a hopper container.
type ContainerHopper struct{}

func (ContainerHopper) Block() world.Block { return hopper{} }
func (ContainerHopper) Type() int          { return protocol.ContainerTypeHopper }
func (ContainerHopper) Size() int          { return 5 }

// ContainerDropper represents a dispenser container.
type ContainerDropper struct{}

func (ContainerDropper) Block() world.Block { return dropper{} }
func (ContainerDropper) Type() int          { return protocol.ContainerTypeDropper }
func (ContainerDropper) Size() int          { return 9 }

// ContainerBarrel represents a barrel container.
type ContainerBarrel struct{}

func (ContainerBarrel) Block() world.Block { return block.NewBarrel() }
func (ContainerBarrel) Type() int          { return protocol.ContainerBarrel }
func (ContainerBarrel) Size() int          { return 27 }

// PlaceFakeContainer places a fake container at the position and world passed.
func PlaceFakeContainer(w *world.World, pos cube.Pos) {
	w.SetBlock(pos, block.NewChest(), nil)
	containerPos = pos
}

// CloseContainer closes the container that the session passed is currently viewing.
func CloseContainer(p *player.Player) {
	menuMu.Lock()
	s := player_session(p)
	m, ok := lastMenus[s]
	if ok {
		if s != session.Nop {
			if closeable, ok := m.submittable.(Closer); ok {
				closeable.Close(p)
			}
			session_writePacket(s, &packet.ContainerClose{
				WindowID:   m.windowID,
				ServerSide: true,
			})

			removeClientSideMenu(p, m)
		}
	}
	menuMu.Unlock()
}

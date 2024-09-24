package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

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

func (ContainerHopper) Block() world.Block { return block.NewHopper() }
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

// ContainerEnderChest represents a enderchest container.
type ContainerEnderChest struct{}

func (ContainerEnderChest) Block() world.Block {
	b := block.NewEnderChest()
	b.Facing = 1
	return b
}
func (ContainerEnderChest) Type() int { return protocol.ContainerTypeContainer }
func (ContainerEnderChest) Size() int { return 27 }

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

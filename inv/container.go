package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
)

var (
	menuMu       sync.Mutex
	lastMenus    = map[block.ContainerViewer]Menu{}
	containerPos cube.Pos
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

type EnderChestContainer struct{}

func (EnderChestContainer) Block() world.Block {
	b := block.NewEnderChest()
	b.Facing = 1
	return b
}
func (EnderChestContainer) Type() int { return protocol.ContainerTypeContainer }
func (EnderChestContainer) Size() int { return 27 }

// lastMenu ...
func lastMenu(v block.ContainerViewer) (Menu, bool) {
	menuMu.Lock()
	defer menuMu.Unlock()
	m, ok := lastMenus[v]
	return m, ok
}

// closeLastMenu ...
func closeLastMenu(p *player.Player, mn Menu) {
	s := player_session(p)
	if s != session.Nop {
		if closeable, ok := mn.submittable.(Closer); ok {
			closeable.Close(p)
		}
		removeClientSideMenu(p, mn)
	}

	menuMu.Lock()
	delete(lastMenus, s)
	menuMu.Unlock()
}

// removeClientSideMenu ...
func removeClientSideMenu(p *player.Player, m Menu) {
	s := player_session(p)
	if s != session.Nop {
		s.ViewBlockUpdate(m.pos, p.World().Block(m.pos), 0)
		airPos := m.pos.Add(cube.Pos{0, 1})
		s.ViewBlockUpdate(airPos, p.World().Block(airPos), 0)
		if c, ok := m.container.(ChestContainer); ok && c.DoubleChest {
			s.ViewBlockUpdate(m.pos.Add(cube.Pos{1, 0, 0}), p.World().Block(m.pos), 0)
			airPos = m.pos.Add(cube.Pos{1, 1})
			s.ViewBlockUpdate(airPos, p.World().Block(airPos), 0)
		}
		delete(lastMenus, s)
	}
}

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

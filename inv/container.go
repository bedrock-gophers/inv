package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
)

var (
	menuMu       sync.Mutex
	lastMenus    = map[block.ContainerViewer]Menu{}
	containerPos cube.Pos
)

func lastMenu(v block.ContainerViewer) (Menu, bool) {
	menuMu.Lock()
	defer menuMu.Unlock()
	m, ok := lastMenus[v]
	return m, ok
}

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

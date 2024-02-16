package inv

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	_ "unsafe"
)

type conn struct {
	session.Conn
	closed bool
}

func (c *conn) ReadPacket() (packet.Packet, error) {
	if c.closed {
		return nil, fmt.Errorf("conn closed")
	}
	return &packet.ActorEvent{}, nil
}

func RedirectPlayerPackets(p *player.Player) {
	s := player_session(p)

	c := fetchPrivateField[session.Conn](s, "conn")
	cn := &conn{c, false}
	updatePrivateField[session.Conn](s, "conn", cn)

	go func() {
		defer func() {
			cn.closed = true
		}()

		for {
			pkt, err := c.ReadPacket()
			if err != nil {
				return
			}
			switch pk := pkt.(type) {
			case *packet.ContainerClose:
				mn, ok := openedMenu(s)
				if ok && pk.WindowID == mn.windowID {
					if closer, ok := mn.submittable.(Closer); ok {
						closer.Close(p)
					}
					s.ViewBlockUpdate(mn.pos, p.World().Block(mn.pos), 0)

					menuMu.Lock()
					delete(openedMenus, s)
					menuMu.Unlock()
				}
			}

			if session_handlePacket(s, pkt) != nil {
				return
			}
		}
	}()
}

// noinspection ALL
//
//go:linkname session_handlePacket github.com/df-mc/dragonfly/server/session.(*Session).handlePacket
func session_handlePacket(s *session.Session, pk packet.Packet) error

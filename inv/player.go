package inv

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sirupsen/logrus"
	_ "unsafe"
)

type conn struct {
	session.Conn
	c chan struct{}
}

func (c *conn) ReadPacket() (packet.Packet, error) {
	<-c.c
	return nil, fmt.Errorf("conn closed")
}

func RedirectPlayerPackets(p *player.Player) {
	s := player_session(p)

	c := fetchPrivateField[session.Conn](s, "conn")
	cn := &conn{c, make(chan struct{})}
	updatePrivateField[session.Conn](s, "conn", cn)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.Panic(err)
			}
			cn.c <- struct{}{}
		}()

		for {
			pkt, err := c.ReadPacket()
			if err != nil {
				return
			}
			switch pk := pkt.(type) {
			case *packet.ContainerClose:
				mn, ok := lastMenu(s)
				if ok && pk.WindowID == mn.windowID {
					if closer, ok := mn.submittable.(Closer); ok {
						closer.Close(p)
					}
					s.ViewBlockUpdate(mn.pos, p.World().Block(mn.pos), 0)
				}
			}
			if s == nil || s == session.Nop {
				return
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

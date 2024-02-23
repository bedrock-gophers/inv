package inv

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/getsentry/sentry-go"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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
				sentry.CurrentHub().Clone().Recover(err)
			}
			cn.c <- struct{}{}
		}()

		for {
			pkt, err := c.ReadPacket()
			if err != nil {
				return
			}
			switch pk := pkt.(type) {
			case *packet.ItemStackRequest:
				for _, data := range pk.Requests {
					for _, action := range data.Actions {
						switch act := action.(type) {
						case *protocol.TakeStackRequestAction:
							if _, ok := lastMenu(s); ok {
								act.Source.ContainerID = 7
							}
						case *protocol.PlaceStackRequestAction:
							if _, ok := lastMenu(s); ok {
								act.Source.ContainerID = 7
							}
						case *protocol.DropStackRequestAction:
							if _, ok := lastMenu(s); ok {
								act.Source.ContainerID = 7
							}
						case *protocol.SwapStackRequestAction:
							if _, ok := lastMenu(s); ok {
								act.Source.ContainerID = 7
							}
						}
					}
				}
			case *packet.ContainerClose:
				mn, ok := lastMenu(s)
				if ok && pk.WindowID == mn.windowID {
					if closer, ok := mn.submittable.(Closer); ok {
						closer.Close(p)
					}
					removeClientSideMenu(p, mn)
				}
			}
			if s == nil || s == session.Nop || pkt == nil {
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

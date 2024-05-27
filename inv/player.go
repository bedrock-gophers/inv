package inv

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"runtime/debug"
	_ "unsafe"
)

type conn struct {
	session.Conn
	c chan struct{}
}

func (c *conn) ReadPacket() (packet.Packet, error) {
	<-c.c
	return nil, fmt.Errorf("connection closed (github.com/bedrock-gophers/inv)")
}

func RedirectPlayerPackets(p *player.Player, recovery func()) {
	s := player_session(p)

	c := fetchPrivateField[session.Conn](s, "conn")
	cn := &conn{c, make(chan struct{})}
	updatePrivateField[session.Conn](s, "conn", cn)

	go func() {
		defer func() {
			cn.c <- struct{}{}

			if recovery != nil {
				if err := recover(); err != nil {
					fmt.Println(err)
					fmt.Println(string(debug.Stack()))
					recovery()
				}
			}
		}()

		for {
			pkt, err := c.ReadPacket()
			if err != nil {
				return
			}

			switch pk := pkt.(type) {
			case *packet.ItemStackRequest:
				handleItemStackRequest(s, pk.Requests)
			case *packet.ContainerClose:
				handleContainerClose(s, p, pk.WindowID)
			}

			if session_handlePacket(s, pkt) != nil {
				return
			}
		}
	}()
}

func handleContainerClose(s *session.Session, p *player.Player, windowID byte) {
	mn, ok := lastMenu(s)
	if ok && windowID == mn.windowID {
		if closer, ok := mn.submittable.(Closer); ok {
			closer.Close(p)
		}
		removeClientSideMenu(p, mn)
	}
}

func handleItemStackRequest(s *session.Session, req []protocol.ItemStackRequest) {
	for _, data := range req {
		for _, action := range data.Actions {
			updateActionContainerID(action, s)
		}
	}
}

// updateActionContainerID updates the container ID of the given action based on the current menu state.
// It is useful in case we use some unimplemented blocks such as hoppers.
func updateActionContainerID(action protocol.StackRequestAction, s *session.Session) {
	switch act := action.(type) {
	case *protocol.TakeStackRequestAction:
		if act.Source.ContainerID != act.Destination.ContainerID || act.Source.ContainerID == protocol.ContainerCursor || act.Source.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.ContainerID = protocol.ContainerLevelEntity
		}
	case *protocol.PlaceStackRequestAction:
		if act.Source.ContainerID != act.Destination.ContainerID || act.Source.ContainerID == protocol.ContainerCursor || act.Source.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.ContainerID = protocol.ContainerLevelEntity
		}
	case *protocol.DropStackRequestAction:
		if act.Source.ContainerID == protocol.ContainerInventory || act.Source.ContainerID == protocol.ContainerCursor || act.Source.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.ContainerID = protocol.ContainerLevelEntity

		}
	case *protocol.SwapStackRequestAction:
		if act.Source.ContainerID != act.Destination.ContainerID || act.Source.ContainerID == protocol.ContainerCursor || act.Source.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.ContainerID = protocol.ContainerLevelEntity
		}
	}
}

// noinspection ALL
//
//go:linkname session_handlePacket github.com/df-mc/dragonfly/server/session.(*Session).handlePacket
func session_handlePacket(s *session.Session, pk packet.Packet) error

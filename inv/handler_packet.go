package inv

import (
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/bedrock-gophers/unsafe/unsafe"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func init() {
	intercept.Hook(packetHandler{})
}

type packetHandler struct{}

func (h packetHandler) HandleClientPacket(ctx *event.Context, p *player.Player, pk packet.Packet) {
	s := unsafe.Session(p)
	switch pkt := pk.(type) {
	case *packet.ItemStackRequest:
		handleItemStackRequest(s, pkt.Requests)
	case *packet.ContainerClose:
		handleContainerClose(ctx, s, p, pkt.WindowID)
	}
}

func (h packetHandler) HandleServerPacket(ctx *event.Context, p *player.Player, pk packet.Packet) {
	// Do nothing
}

func handleContainerClose(ctx *event.Context, s *session.Session, p *player.Player, windowID byte) {
	mn, ok := lastMenu(s)
	currentID := fetchPrivateField[atomic.Uint32](s, "openedWindowID")
	if ok && windowID == mn.windowID && byte(currentID.Load()) == windowID {
		if closer, ok := mn.submittable.(Closer); ok {
			closer.Close(p)
		}

		if mn.containerClose != nil {
			mn.containerClose(mn.inventory)
		}

		removeClientSideMenu(p, mn)
	} else {
		p.OpenBlockContainer(mn.pos)
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

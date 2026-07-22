package inv

import (
	"context"
	"sync/atomic"

	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/bedrock-gophers/unsafe/unsafe"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func init() {
	intercept.Hook(packetHandler{})
}

type packetHandler struct{}

func (h packetHandler) HandleClientPacket(ctx *intercept.Context, pk packet.Packet) {
	switch pk.(type) {
	case *packet.ItemStackRequest, *packet.ContainerClose:
	default:
		return
	}

	ha, ok := ctx.Val().Handle()
	if !ok {
		return
	}
	_, _ = player.Call(context.Background(), ha, func(tx *world.Tx, p *player.Player) (struct{}, error) {
		s := unsafe.Session(p)
		switch pkt := pk.(type) {
		case *packet.ItemStackRequest:
			handleItemStackRequest(s, pkt.Requests)
		case *packet.ContainerClose:
			handleContainerClose(ctx, p, s, pkt.WindowID)
		}
		return struct{}{}, nil
	})
}

func (h packetHandler) HandleServerPacket(_ *intercept.Context, _ packet.Packet) {
	// Do nothing
}

func handleContainerClose(ctx *intercept.Context, p *player.Player, s *session.Session, windowID byte) {
	mn, ok := lastMenu(s)
	if !ok {
		return
	}
	currentID := fetchPrivateField[atomic.Uint32](s, "openedWindowID")
	if byte(currentID.Load()) == windowID && windowID == mn.windowID {
		closeLastMenu(p, mn)
		return
	}
	ctx.Cancel()
	p.OpenBlockContainer(mn.pos, p.Tx())
	closeLastMenu(p, mn)
}

func handleItemStackRequest(s *session.Session, req []protocol.ItemStackRequest) {
	_, anvil := anvilMenu(s)
	for i := range req {
		actions := req[i].Actions[:0]
		for _, action := range req[i].Actions {
			if anvil {
				switch action.(type) {
				case *protocol.CraftRecipeOptionalStackRequestAction, *protocol.CreateStackRequestAction:
					continue
				}
			}
			updateActionContainerID(action, s)
			actions = append(actions, action)
		}
		req[i].Actions = actions
	}
}

// updateActionContainerID updates the container ID of the given action based on the current menu state.
// It is useful in case we use some unimplemented blocks such as hoppers.
func updateActionContainerID(action protocol.StackRequestAction, s *session.Session) {
	if m, ok := anvilMenu(s); ok {
		switch act := action.(type) {
		case *protocol.TakeStackRequestAction:
			rewriteAnvilSlot(&act.Source, m)
			rewriteAnvilSlot(&act.Destination, m)
		case *protocol.PlaceStackRequestAction:
			rewriteAnvilSlot(&act.Source, m)
			rewriteAnvilSlot(&act.Destination, m)
		case *protocol.DropStackRequestAction:
			rewriteAnvilSlot(&act.Source, m)
		case *protocol.DestroyStackRequestAction:
			rewriteAnvilSlot(&act.Source, m)
		case *protocol.SwapStackRequestAction:
			rewriteAnvilSlot(&act.Source, m)
			rewriteAnvilSlot(&act.Destination, m)
		}
		return
	}

	switch act := action.(type) {
	case *protocol.TakeStackRequestAction:
		if act.Source.Container.ContainerID != act.Destination.Container.ContainerID || act.Source.Container.ContainerID == protocol.ContainerCursor || act.Source.Container.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.Container.ContainerID = protocol.ContainerLevelEntity
		}
	case *protocol.PlaceStackRequestAction:
		if act.Source.Container.ContainerID != act.Destination.Container.ContainerID || act.Source.Container.ContainerID == protocol.ContainerCursor || act.Source.Container.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.Container.ContainerID = protocol.ContainerLevelEntity
		}
	case *protocol.DropStackRequestAction:
		if act.Source.Container.ContainerID == protocol.ContainerInventory || act.Source.Container.ContainerID == protocol.ContainerCursor || act.Source.Container.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.Container.ContainerID = protocol.ContainerLevelEntity

		}
	case *protocol.SwapStackRequestAction:
		if act.Source.Container.ContainerID != act.Destination.Container.ContainerID || act.Source.Container.ContainerID == protocol.ContainerCursor || act.Source.Container.ContainerID == protocol.ContainerHotBar {
			break
		}
		if _, ok := lastMenu(s); ok {
			act.Source.Container.ContainerID = protocol.ContainerLevelEntity
		}
	}
}

func anvilMenu(s *session.Session) (Menu, bool) {
	m, ok := lastMenu(s)
	return m, ok && m.container != nil && m.container.Type() == protocol.ContainerTypeAnvil
}

func rewriteAnvilSlot(slot *protocol.StackRequestSlotInfo, m Menu) {
	switch slot.Container.ContainerID {
	case protocol.ContainerAnvilInput:
		slot.Container.ContainerID = protocol.ContainerLevelEntity
		slot.Slot = 0
	case protocol.ContainerAnvilMaterial:
		slot.Container.ContainerID = protocol.ContainerLevelEntity
		slot.Slot = 1
	case protocol.ContainerAnvilResultPreview, protocol.ContainerCreatedOutput:
		slot.Container.ContainerID = protocol.ContainerLevelEntity
		slot.Slot = 2
	default:
		return
	}
	if slot.StackNetworkID < 0 && m.inventory != nil {
		if stack, err := m.inventory.Item(int(slot.Slot)); err == nil {
			slot.StackNetworkID = item_id(stack)
		}
	}
}

package entityinv

import (
	"bytes"
	"context"

	"github.com/bedrock-gophers/intercept/intercept"
	unsafe2 "github.com/bedrock-gophers/unsafe/unsafe"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func init() {
	nextActorRuntimeID.Store(1 << 62)
	intercept.Hook(packetHandler{})
}

type menuHandler struct {
	inventory.NopHandler
	menu Menu
}

func (h menuHandler) HandleTake(ctx *inventory.Context, _ int, stack item.Stack) {
	p := ctx.Val().(*player.Player)
	ctx.Cancel()
	h.menu.submittable.Submit(p, stack)
}

func (h menuHandler) HandlePlace(ctx *inventory.Context, _ int, _ item.Stack) { ctx.Cancel() }

func (h menuHandler) HandleDrop(ctx *inventory.Context, _ int, stack item.Stack) {
	p := ctx.Val().(*player.Player)
	ctx.Cancel()
	h.menu.submittable.Submit(p, stack)
}

type packetHandler struct{}

func (packetHandler) HandleClientPacket(ctx *intercept.Context, pk packet.Packet) {
	closePacket, ok := pk.(*packet.ContainerClose)
	if !ok {
		return
	}
	h, ok := ctx.Val().Handle()
	if !ok {
		return
	}
	_, _ = player.Call(context.Background(), h, func(_ *world.Tx, p *player.Player) (struct{}, error) {
		s := unsafe2.Session(p)
		m, ok := takeLastMenu(s, &closePacket.WindowID)
		if !ok {
			return struct{}{}, nil
		}
		removeGraphic(s, m)
		closeCallbacks(p, m)
		return struct{}{}, nil
	})
}

func (packetHandler) HandleServerPacket(_ *intercept.Context, pk packet.Packet) {
	identifiers, ok := pk.(*packet.AvailableActorIdentifiers)
	if !ok {
		return
	}
	identifiers.SerialisedEntityIdentifiers = addActorIdentifier(identifiers.SerialisedEntityIdentifiers)
}

type actorIdentifiers struct {
	IDList []actorID `nbt:"idlist"`
}

type actorID struct {
	ID string `nbt:"id"`
}

func addActorIdentifier(data []byte) []byte {
	var actors actorIdentifiers
	if err := nbt.NewDecoder(bytes.NewReader(data)).Decode(&actors); err != nil {
		return data
	}
	for _, actor := range actors.IDList {
		if actor.ID == actorIdentifier {
			return data
		}
	}
	actors.IDList = append(actors.IDList, actorID{ID: actorIdentifier})
	encoded, err := nbt.Marshal(actors)
	if err != nil {
		return data
	}
	return encoded
}

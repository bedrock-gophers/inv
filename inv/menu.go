package inv

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	_ "unsafe"
)

func ShowMenu(p *player.Player, inv *inventory.Inventory, customName string) {
	var bl world.Block
	switch inv.Size() {
	case 27, 54:
		bl = block.Chest{}
	default:
		panic("invalid size")
	}
	s := player_session(p)

	pos := cube.PosFromVec3(p.Rotation().Vec3().Mul(-2).Add(p.Position()))
	s.ViewBlockUpdate(pos, bl, 0)
	s.ViewBlockUpdate(pos.Add(cube.Pos{0, 1}), block.Air{}, 0)

	blockPos := blockPosToProtocol(pos)
	data := createFakeInventoryNBT(customName, inv)
	if inv.Size() == 54 {
		data["x"], data["y"], data["z"] = blockPos.X(), blockPos.Y(), blockPos.Z()

	}
	data["x"], data["y"], data["z"] = blockPos.X(), blockPos.Y(), blockPos.Z()
	session_writePacket(s, &packet.BlockActorData{
		Position: blockPos,
		NBTData:  data,
	})

	nextID := session_nextWindowID(s)
	updatePrivateField(s, "openedPos", *atomic.NewValue(cube.Pos{0, 255, 0}))
	updatePrivateField(s, "openedWindow", *atomic.NewValue(inv))

	updatePrivateField(s, "containerOpened", *atomic.NewBool(true))
	updatePrivateField(s, "openedContainerID", *atomic.NewUint32(uint32(nextID)))

	session_writePacket(s, &packet.ContainerOpen{
		WindowID:                nextID,
		ContainerPosition:       blockPos,
		ContainerType:           0,
		ContainerEntityUniqueID: -1,
	})
	session_sendInv(s, inv, uint32(nextID))
}

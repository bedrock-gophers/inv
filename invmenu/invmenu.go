package invmenu

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
	"time"
	_ "unsafe"
)

var (
	fakeInventoriesOpen = map[*player.Player]*inventory.Inventory{}
	fakeInventoriesMu   sync.Mutex
)

func ShowInventory(p *player.Player, name string, inv *inventory.Inventory) {
	var bl world.Block
	switch inv.Size() {
	case 5:
		bl, _ = world.BlockByName("minecraft:hopper", map[string]any{})
	case 9:
		bl, _ = world.BlockByName("minecraft:dispenser", map[string]any{})
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
	data := createFakeInventoryNBT(name, inv, pos)
	if inv.Size() == 54 {
		data["x"], data["y"], data["z"] = blockPos.X(), blockPos.Y(), blockPos.Z()

	}
	data["x"], data["y"], data["z"] = blockPos.X(), blockPos.Y(), blockPos.Z()
	session_writePacket(s, &packet.BlockActorData{
		Position: blockPos,
		NBTData:  data,
	})

	time.AfterFunc(time.Millisecond*50, func() {
		nextID := session_nextWindowID(s)
		session_updateOpenedPos(s, pos)
		session_updateOpenedWindow(s, inv)

		fakeInventoriesMu.Lock()
		fakeInventoriesOpen[p] = inv
		fakeInventoriesMu.Unlock()

		session_updateContainerOpenedState(s, true)
		session_updateOpenedContainerID(s, uint32(nextID))
		session_writePacket(s, &packet.ContainerOpen{
			WindowID:                nextID,
			ContainerPosition:       blockPos,
			ContainerType:           0,
			ContainerEntityUniqueID: -1,
		})
		session_sendInv(s, inv, uint32(nextID))
	})
}

// blockPosToProtocl converts a cube.Pos to a protocol.BlockPos.
func blockPosToProtocol(pos cube.Pos) protocol.BlockPos {
	return protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}
}

// createFakeInventoryNBT creates a map of NBT data for a fake inventory with the name passed and the inventory
func createFakeInventoryNBT(name string, inv *inventory.Inventory, pos cube.Pos) map[string]interface{} {
	m := map[string]interface{}{"CustomName": name}
	switch inv.Size() {
	case 27, 54:
		m["id"] = "Chest"
	}
	return m
}

// noinspection ALL
//
//go:linkname player_session github.com/df-mc/dragonfly/server/player.(*Player).session
func player_session(*player.Player) *session.Session

// noinspection ALL
//
//go:linkname session_writePacket github.com/df-mc/dragonfly/server/session.(*Session).writePacket
func session_writePacket(*session.Session, packet.Packet)

// noinspection ALL
//
//go:linkname session_nextWindowID github.com/df-mc/dragonfly/server/session.(*Session).nextWindowID
func session_nextWindowID(*session.Session) byte

// noinspection ALL
//
//go:linkname session_updateOpenedPos github.com/df-mc/dragonfly/server/session.(*Session).updateOpenedPos
func session_updateOpenedPos(*session.Session, cube.Pos)

// noinspection ALL
//
//go:linkname session_updateOpenedWindow github.com/df-mc/dragonfly/server/session.(*Session).updateOpenedWindow
func session_updateOpenedWindow(*session.Session, *inventory.Inventory)

// noinspection ALL
//
//go:linkname session_updateContainerOpenedState github.com/df-mc/dragonfly/server/session.(*Session).updateContainerOpenedState
func session_updateContainerOpenedState(*session.Session, bool)

// noinspection ALL
//
//go:linkname session_updateOpenedContainerID github.com/df-mc/dragonfly/server/session.(*Session).updateOpenedContainerID
func session_updateOpenedContainerID(*session.Session, uint32)

// noinspection ALL
//
//go:linkname session_sendInv github.com/df-mc/dragonfly/server/session.(*Session).sendInv
func session_sendInv(*session.Session, *inventory.Inventory, uint32)

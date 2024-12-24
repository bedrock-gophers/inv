package inv

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Menu is a menu that can be sent to a player. It can be used to display a custom inventory to a player.
type Menu struct {
	name      string
	container Container

	inventory      *inventory.Inventory
	containerClose func(inv *inventory.Inventory)

	submittable Submittable

	pos cube.Pos

	windowID byte
	custom   bool
}

// NewMenu creates a new menu with the submittable passed, the name passed and the container passed.
func NewMenu(submittable Submittable, name string, container Container) Menu {
	return Menu{name: name, submittable: submittable, container: container, inventory: inventory.New(container.Size(), func(slot int, before, after item.Stack) {})}
}

// NewCustomMenu creates a new menu with the name, container and inventory.
func NewCustomMenu(name string, container Container, inv *inventory.Inventory, containerClose func(inv *inventory.Inventory)) Menu {
	return Menu{name: name, container: container, inventory: inv, containerClose: containerClose, custom: true}
}

// WithStacks sets the stacks of the menu to the stacks passed.
func (m Menu) WithStacks(stacks ...item.Stack) Menu {
	m.inventory.Clear()
	for i, it := range stacks {
		_ = m.inventory.SetItem(i, it)
	}
	return m
}

// Submittable is a type that can be implemented by a Menu to be called when a menu is submitted.
type Submittable interface {
	Submit(p *player.Player, it item.Stack)
}

// Closer is a type that can be implemented by a Submittable to be called when a menu is closed.
type Closer interface {
	Close(p *player.Player)
}

// SendMenu sends a menu to a player. The menu passed will be displayed to the player
func SendMenu(p *player.Player, m Menu) {
	sendMenu(p, m, false)
}

// UpdateMenu updates the menu that the player passed is currently viewing to the menu passed.
func UpdateMenu(p *player.Player, m Menu) {
	sendMenu(p, m, true)
}

// sendMenu sends the menu to a player.
func sendMenu(p *player.Player, m Menu, update bool) {
	s := player_session(p)

	if !m.custom {
		m.inventory.Handle(handler{menu: m})
	}

	pos := cube.PosFromVec3(p.Rotation().Vec3().Mul(-1.5).Add(p.Position()))
	blockPos := blockPosToProtocol(pos)

	var nextID byte
	if update {
		mn, ok := lastMenu(s)
		if ok {
			pos = mn.pos
			nextID = mn.windowID

			if c, ok := mn.container.(ContainerChest); ok && c.DoubleChest {
				s.ViewBlockUpdate(pos.Add(cube.Pos{1, 0, 0}), block.Air{}, 0)
				s.ViewBlockUpdate(pos.Add(cube.Pos{1, 1}), block.Air{}, 0)
			}
		}
	} else {
		if m, ok := lastMenu(s); ok && m.pos != pos {
			closeLastMenu(p, m)
		}
		nextID = session_nextWindowID(s)
	}
	s.ViewBlockUpdate(pos, m.container.Block(), 0)
	s.ViewBlockUpdate(pos.Add(cube.Pos{0, 1}), block.Air{}, 0)

	data := createFakeInventoryNBT(m.name, m.container)

	if m.container.Size() == 54 {
		s.ViewBlockUpdate(pos.Add(cube.Pos{1, 0, 0}), m.container.Block(), 0)
		s.ViewBlockUpdate(pos.Add(cube.Pos{1, 1}), block.Air{}, 0)

		data["pairz"] = int32(pos[2])
		data["pairx"] = int32(pos[0] + 1)
	}

	session_writePacket(s, &packet.BlockActorData{
		Position: blockPos,
		NBTData:  data,
	})

	posPtr := atomic.Pointer[cube.Pos]{}
	invPtr := atomic.Pointer[inventory.Inventory]{}
	containerOpenedPtr := atomic.Bool{}
	openedContainerIdPtr := atomic.Uint32{}
	openedWindowIdPtr := atomic.Uint32{}

	posPtr.Store(&pos)
	invPtr.Store(m.inventory)
	containerOpenedPtr.Store(true)
	openedContainerIdPtr.Store(uint32(nextID))
	openedWindowIdPtr.Store(uint32(nextID))

	updatePrivateField(s, "openedPos", posPtr)
	updatePrivateField(s, "openedWindow", invPtr)

	updatePrivateField(s, "containerOpened", containerOpenedPtr)
	updatePrivateField(s, "openedContainerID", openedContainerIdPtr)
	updatePrivateField(s, "openedWindowID", openedWindowIdPtr)

	time.AfterFunc(time.Millisecond*50, func() {
		session_writePacket(s, &packet.ContainerOpen{
			WindowID:                nextID,
			ContainerPosition:       blockPos,
			ContainerType:           byte(m.container.Type()),
			ContainerEntityUniqueID: -1,
		})
		session_sendInv(s, m.inventory, uint32(nextID))
	})

	m.pos = pos
	m.windowID = nextID

	menuMu.Lock()
	lastMenus[s] = m
	menuMu.Unlock()
}

var (
	menuMu    sync.Mutex
	lastMenus = map[*session.Session]Menu{}
)

func lastMenu(s *session.Session) (Menu, bool) {
	menuMu.Lock()
	defer menuMu.Unlock()
	m, ok := lastMenus[s]
	return m, ok
}

func closeLastMenu(p *player.Player, mn Menu) {
	s := player_session(p)
	if s != session.Nop {
		if closeable, ok := mn.submittable.(Closer); ok {
			closeable.Close(p)
		}
		if mn.containerClose != nil {
			mn.containerClose(mn.inventory)
		}
		removeClientSideMenu(p, mn)
	}

	menuMu.Lock()
	delete(lastMenus, s)
	menuMu.Unlock()
}

func removeClientSideMenu(p *player.Player, m Menu) {
	s := player_session(p)
	if s != session.Nop {
		s.ViewBlockUpdate(m.pos, p.Tx().Block(m.pos), 0)
		airPos := m.pos.Add(cube.Pos{0, 1})
		s.ViewBlockUpdate(airPos, p.Tx().Block(airPos), 0)
		if c, ok := m.container.(ContainerChest); ok && c.DoubleChest {
			s.ViewBlockUpdate(m.pos.Add(cube.Pos{1, 0, 0}), p.Tx().Block(m.pos), 0)
			airPos = m.pos.Add(cube.Pos{1, 1})
			s.ViewBlockUpdate(airPos, p.Tx().Block(airPos), 0)
		}
		delete(lastMenus, s)
	}
}

// blockPosToProtocol converts a cube.Pos to a protocol.BlockPos.
func blockPosToProtocol(pos cube.Pos) protocol.BlockPos {
	return protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}
}

// createFakeInventoryNBT creates a map of NBT data for a fake inventory with the name passed and the inventory
func createFakeInventoryNBT(name string, container Container) map[string]interface{} {
	m := map[string]interface{}{"CustomName": name}
	switch container.Type() {
	case protocol.ContainerTypeContainer:
		m["id"] = "Chest"
	case protocol.ContainerTypeHopper:
		m["id"] = "Hopper"
	case protocol.ContainerTypeDropper:
		m["id"] = "Dropper"
	default:
		panic("should never happen")
	}
	return m
}

// updatePrivateField sets a private field of a session to the value passed.
func updatePrivateField[T any](s *session.Session, name string, value T) {
	reflectedValue := reflect.ValueOf(s).Elem()
	privateFieldValue := reflectedValue.FieldByName(name)

	privateFieldValue = reflect.NewAt(privateFieldValue.Type(), unsafe.Pointer(privateFieldValue.UnsafeAddr())).Elem()

	privateFieldValue.Set(reflect.ValueOf(value))
}

// fetchPrivateField fetches a private field of a session.
func fetchPrivateField[T any](s *session.Session, name string) T {
	reflectedValue := reflect.ValueOf(s).Elem()
	privateFieldValue := reflectedValue.FieldByName(name)
	privateFieldValue = reflect.NewAt(privateFieldValue.Type(), unsafe.Pointer(privateFieldValue.UnsafeAddr())).Elem()

	return privateFieldValue.Interface().(T)
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
//go:linkname session_sendInv github.com/df-mc/dragonfly/server/session.(*Session).sendInv
func session_sendInv(*session.Session, *inventory.Inventory, uint32)

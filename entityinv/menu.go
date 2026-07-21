// Package entityinv provides arbitrary-sized inventory menus backed by a client-side entity.
package entityinv

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
	_ "unsafe"

	unsafe2 "github.com/bedrock-gophers/unsafe/unsafe"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const actorIdentifier = "inventoryui:inventoryui"

var nextActorRuntimeID atomic.Uint64

// Menu is an arbitrary-sized inventory menu displayed using an invisible client-side entity.
type Menu struct {
	name           string
	inventory      *inventory.Inventory
	containerClose func(*inventory.Inventory)
	submittable    Submittable
	custom         bool

	windowID       byte
	actorRuntimeID uint64
}

// NewMenu creates an entity inventory menu with size slots. Size must be greater than zero.
func NewMenu(submittable Submittable, name string, size int) Menu {
	return Menu{
		name:        name,
		submittable: submittable,
		inventory:   inventory.New(size, nil),
	}
}

// NewCustomMenu creates an entity inventory menu backed by inv. containerClose is called when the menu closes.
func NewCustomMenu(name string, inv *inventory.Inventory, containerClose func(*inventory.Inventory)) Menu {
	if inv == nil {
		panic("entityinv: nil inventory")
	}
	return Menu{name: name, inventory: inv, containerClose: containerClose, custom: true}
}

// Inventory returns the inventory backing the menu.
func (m Menu) Inventory() *inventory.Inventory { return m.inventory }

// WithStacks replaces the stacks displayed by the menu. Extra stacks are ignored.
func (m Menu) WithStacks(stacks ...item.Stack) Menu {
	m.inventory.Clear()
	for slot, stack := range stacks {
		if slot >= m.inventory.Size() {
			break
		}
		_ = m.inventory.SetItem(slot, stack)
	}
	return m
}

// Submittable handles a stack selected from a Menu.
type Submittable interface {
	Submit(p *player.Player, stack item.Stack)
}

// Closer may be implemented by a Submittable to handle the menu closing.
type Closer interface {
	Close(p *player.Player)
}

// SendMenu opens m for p.
func SendMenu(p *player.Player, m Menu) { sendMenu(p, m, false) }

// UpdateMenu replaces the contents and handlers of p's open entity inventory. If none is open, it opens m.
// The existing title and size remain visible until the menu is reopened.
func UpdateMenu(p *player.Player, m Menu) { sendMenu(p, m, true) }

func sendMenu(p *player.Player, m Menu, update bool) {
	s := unsafe2.Session(p)
	if s == session.Nop {
		return
	}
	if !m.custom {
		m.inventory.Handle(menuHandler{menu: m})
	}

	if current, ok := lastMenu(s); update && ok {
		m.windowID = current.windowID
		m.actorRuntimeID = current.actorRuntimeID
		openSessionInventory(s, m.inventory, m.windowID)
		storeLastMenu(s, m)
		session_sendInv(s, m.inventory, uint32(m.windowID))
		return
	}

	if current, ok := takeLastMenu(s, nil); ok {
		removeGraphic(s, current)
		closeCallbacks(p, current)
	}

	m.windowID = session_nextWindowID(s)
	m.actorRuntimeID = nextActorRuntimeID.Add(1)
	openSessionInventory(s, m.inventory, m.windowID)
	storeLastMenu(s, m)

	pos := p.Position()
	metadata := protocol.NewEntityMetadata()
	metadata[protocol.EntityDataKeyContainerType] = byte(0xff)
	metadata[protocol.EntityDataKeyContainerSize] = int32(m.inventory.Size())
	metadata[protocol.EntityDataKeyName] = encodedTitle(m.name, m.inventory.Size())

	session_writePacket(s, &packet.AddActor{
		EntityUniqueID:  int64(m.actorRuntimeID),
		EntityRuntimeID: m.actorRuntimeID,
		EntityType:      actorIdentifier,
		Position:        mgl32.Vec3{float32(pos[0]), float32(pos[1]), float32(pos[2])},
		EntityMetadata:  metadata,
	})
	session_writePacket(s, &packet.ContainerOpen{
		WindowID:                m.windowID,
		ContainerType:           byte(protocol.ContainerTypeContainer),
		ContainerPosition:       protocol.BlockPos{},
		ContainerEntityUniqueID: int64(m.actorRuntimeID),
	})
	session_sendInv(s, m.inventory, uint32(m.windowID))
}

// CloseContainer closes p's open entity inventory, if any.
func CloseContainer(p *player.Player) {
	s := unsafe2.Session(p)
	m, ok := takeLastMenu(s, nil)
	if !ok {
		return
	}
	resetSessionInventory(s)
	session_writePacket(s, &packet.ContainerClose{
		WindowID:      m.windowID,
		ContainerType: byte(protocol.ContainerTypeContainer),
		ServerSide:    true,
	})
	removeGraphic(s, m)
	closeCallbacks(p, m)
}

func encodedTitle(name string, size int) string {
	rows := (size + 8) / 9
	if rows > 6 {
		rows = 6
	}
	scroll := 0
	if rows*9 < size {
		scroll = 1
	}
	return "§" + string(rune('0'+rows)) + "§" + string(rune('0'+scroll)) + strings.Repeat("§r", 10) + name
}

func openSessionInventory(s *session.Session, inv *inventory.Inventory, windowID byte) {
	pos := cube.Pos{}
	sessionField[atomic.Pointer[cube.Pos]](s, "openedPos").Store(&pos)
	sessionField[atomic.Pointer[inventory.Inventory]](s, "openedWindow").Store(inv)
	sessionField[atomic.Bool](s, "containerOpened").Store(true)
	sessionField[atomic.Uint32](s, "openedContainerID").Store(uint32(protocol.ContainerTypeContainer))
	sessionField[atomic.Uint32](s, "openedWindowID").Store(uint32(windowID))
}

func resetSessionInventory(s *session.Session) {
	pos := cube.Pos{}
	sessionField[atomic.Pointer[cube.Pos]](s, "openedPos").Store(&pos)
	sessionField[atomic.Pointer[inventory.Inventory]](s, "openedWindow").Store(inventory.New(1, nil))
	sessionField[atomic.Bool](s, "containerOpened").Store(false)
	sessionField[atomic.Uint32](s, "openedContainerID").Store(0)
}

func removeGraphic(s *session.Session, m Menu) {
	session_writePacket(s, &packet.RemoveActor{EntityUniqueID: int64(m.actorRuntimeID)})
}

func closeCallbacks(p *player.Player, m Menu) {
	if closeable, ok := m.submittable.(Closer); ok {
		closeable.Close(p)
	}
	if m.containerClose != nil {
		m.containerClose(m.inventory)
	}
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

func storeLastMenu(s *session.Session, m Menu) {
	menuMu.Lock()
	lastMenus[s] = m
	menuMu.Unlock()
}

func takeLastMenu(s *session.Session, windowID *byte) (Menu, bool) {
	menuMu.Lock()
	defer menuMu.Unlock()
	m, ok := lastMenus[s]
	if !ok || windowID != nil && m.windowID != *windowID {
		return Menu{}, false
	}
	delete(lastMenus, s)
	return m, true
}

func sessionField[T any](s *session.Session, name string) *T {
	v := reflect.ValueOf(s).Elem().FieldByName(name)
	return (*T)(unsafe.Pointer(v.UnsafeAddr()))
}

//go:linkname session_writePacket github.com/df-mc/dragonfly/server/session.(*Session).writePacket
func session_writePacket(*session.Session, packet.Packet)

//go:linkname session_nextWindowID github.com/df-mc/dragonfly/server/session.(*Session).nextWindowID
func session_nextWindowID(*session.Session) byte

//go:linkname session_sendInv github.com/df-mc/dragonfly/server/session.(*Session).sendInv
func session_sendInv(*session.Session, *inventory.Inventory, uint32)

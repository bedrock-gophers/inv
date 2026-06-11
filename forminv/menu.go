// Package forminv provides inventory-shaped Bedrock form menus.
package forminv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
)

const (
	// TitlePrefix is the formatting prefix used by the bundled resource pack to route a normal menu form to the
	// inventory-shaped chest renderer. Titles passed to NewMenu are prefixed automatically.
	TitlePrefix = "§c"

	// SmallChestSlots is the number of buttons shown by a single chest layout.
	SmallChestSlots = 27
	// LargeChestSlots is the number of buttons shown by a double chest layout.
	LargeChestSlots = 54
)

// Button is a form button shown in an inventory slot.
type Button = form.Button

// NewButton creates a button with text and an optional texture path or URL.
func NewButton(text, image string) Button { return form.NewButton(text, image) }

// Slot is a single form inventory slot. Button controls how the slot is rendered. Value is returned unchanged
// when the slot is submitted.
type Slot struct {
	Button form.Button
	Value  any

	index *int
}

// NewSlot creates a slot positioned at the index passed.
func NewSlot(index int, text, image string, value any) Slot {
	return Slot{Button: form.NewButton(text, image), Value: value}.At(index)
}

// At returns a copy of the slot positioned at the index passed.
func (s Slot) At(index int) Slot {
	s.index = &index
	return s
}

// Container represents a form inventory layout.
type Container interface {
	Size() int
}

// ContainerChest represents a chest form inventory. It can be a single chest or a double chest.
type ContainerChest struct{ DoubleChest bool }

// Size returns the amount of slots in the chest form inventory.
func (c ContainerChest) Size() int {
	if c.DoubleChest {
		return LargeChestSlots
	}
	return SmallChestSlots
}

// Menu is an inventory-shaped form menu that can be sent to a player.
type Menu struct {
	name        string
	container   Container
	submittable Submittable
	slots       []Slot
}

// NewMenu creates a form inventory menu with the submittable, name and container passed.
func NewMenu(submittable Submittable, name string, container Container) Menu {
	if container == nil {
		panic("forminv: nil container")
	}
	if size := container.Size(); size <= 0 {
		panic(fmt.Sprintf("forminv: invalid container size %d", size))
	}
	return Menu{name: name, submittable: submittable, container: container}
}

// NewChestMenu creates a double chest form inventory menu.
func NewChestMenu(submittable Submittable, name string) Menu {
	return NewMenu(submittable, name, ContainerChest{DoubleChest: true})
}

// Submittable is implemented by form inventory menus to handle clicked buttons.
type Submittable interface {
	Submit(p *player.Player, slot Slot)
}

// Closer may be implemented by a Submittable to handle the player closing the form inventory.
type Closer interface {
	Close(p *player.Player)
}

// WithButtons sets the buttons shown in the menu.
//
// Unpositioned buttons fill the first empty slots. Use WithSlots and Slot.At or NewSlotAt to place slots at
// specific indexes. Empty slots are filled with placeholder buttons so the client renders a stable inventory
// grid.
func (m Menu) WithButtons(buttons ...form.Button) Menu {
	slots := make([]Slot, len(buttons))
	for i, button := range buttons {
		slots[i] = Slot{Button: button, Value: button}
	}
	return m.WithSlots(slots...)
}

// WithSlots sets the slots shown in the menu.
//
// Use Slot.At or NewSlotAt to place slots at specific indexes. Unpositioned slots fill the first empty slots.
// Empty slots are filled with placeholder buttons so the client renders a stable inventory grid.
func (m Menu) WithSlots(slots ...Slot) Menu {
	m.slots = layoutSlots(m.container.Size(), slots)
	return m
}

// Form returns the underlying Dragonfly form. It may be passed directly to player.SendForm when needed.
func (m Menu) Form() Form {
	slots := m.slots
	if slots == nil {
		slots = layoutSlots(m.container.Size(), nil)
	}
	m.slots = slots
	return Form{menu: m}
}

// SendMenu sends a form inventory menu to a player.
func SendMenu(p *player.Player, m Menu) { p.SendForm(m.Form()) }

// UpdateMenu resends a form inventory menu to a player.
func UpdateMenu(p *player.Player, m Menu) { SendMenu(p, m) }

// Form is the Dragonfly form implementation produced by a Menu.
type Form struct {
	menu Menu
}

// MarshalJSON marshals the form inventory as a normal Bedrock menu form. The bundled resource pack renders it
// as an inventory grid by matching the title prefix.
func (f Form) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "form",
		"title":   f.Title(),
		"content": "",
		"buttons": f.Buttons(),
	})
}

// SubmitJSON handles the selected form button index.
func (f Form) SubmitJSON(b []byte, submitter form.Submitter, tx *world.Tx) error {
	if b == nil {
		f.close(submitter)
		return nil
	}
	var index uint
	if err := json.Unmarshal(b, &index); err != nil {
		return fmt.Errorf("cannot parse button index as int: %w", err)
	}
	if index >= uint(len(f.menu.slots)) {
		return fmt.Errorf("button index points to inexistent slot: %v (only %v slots present)", index, len(f.menu.slots))
	}
	slot := f.menu.slots[index]
	if slot.Value == nil {
		return nil
	}
	if p, ok := submitter.(*player.Player); ok {
		f.menu.submittable.Submit(p, slot)
	}
	return nil
}

// Title returns the rendered form title.
func (f Form) Title() string {
	return TitlePrefix + f.menu.name
}

// Buttons returns the rendered form buttons.
func (f Form) Buttons() []form.Button {
	buttons := make([]form.Button, len(f.menu.slots))
	for i, slot := range f.menu.slots {
		buttons[i] = slot.Button
	}
	return buttons
}

func (f Form) close(submitter form.Submitter) {
	if p, ok := submitter.(*player.Player); ok {
		if closeable, ok := f.menu.submittable.(Closer); ok {
			closeable.Close(p)
		}
	}
}

func layoutSlots(size int, input []Slot) []Slot {
	slots := make([]Slot, size)
	used := make([]bool, size)

	var unpositioned []Slot
	for _, slot := range input {
		if slot.index != nil && *slot.index >= 0 && *slot.index < size {
			slots[*slot.index] = slot
			used[*slot.index] = true
			continue
		}
		if idx, text, ok := parseSlotPrefix(slot.Button.Text); ok && idx >= 0 && idx < size {
			slot.Button.Text = text
			slots[idx] = slot
			used[idx] = true
			continue
		}
		unpositioned = append(unpositioned, slot)
	}

	slotIdx := 0
	for _, slot := range unpositioned {
		for slotIdx < size && used[slotIdx] {
			slotIdx++
		}
		if slotIdx >= size {
			break
		}
		slots[slotIdx] = slot
		used[slotIdx] = true
		slotIdx++
	}

	for i := range slots {
		if !used[i] {
			slots[i] = Slot{Button: form.NewButton("", "")}
		}
	}
	return slots
}

func parseSlotPrefix(text string) (int, string, bool) {
	if !strings.HasPrefix(text, "§") {
		return 0, text, false
	}
	rest := text[len("§"):]
	end := strings.Index(rest, "§")
	if end < 0 {
		return 0, text, false
	}
	idx, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0, text, false
	}
	return idx, rest[end+len("§"):], true
}

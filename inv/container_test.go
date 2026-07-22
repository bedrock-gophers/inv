package inv

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestContainerAnvil(t *testing.T) {
	c := ContainerAnvil{}
	if _, ok := c.Block().(block.Anvil); !ok {
		t.Fatalf("Block() returned %T, want block.Anvil", c.Block())
	}
	if got := c.Type(); got != protocol.ContainerTypeAnvil {
		t.Fatalf("Type() returned %d, want %d", got, protocol.ContainerTypeAnvil)
	}
	if got := c.Size(); got != 3 {
		t.Fatalf("Size() returned %d, want 3", got)
	}
}

func TestContainerAnvilHasNoBlockActorData(t *testing.T) {
	if got := createFakeInventoryNBT("Anvil", ContainerAnvil{}); got != nil {
		t.Fatalf("createFakeInventoryNBT() returned %v, want nil", got)
	}
}

func TestContainerAnvilRewritesSpecialSlots(t *testing.T) {
	s := &session.Session{}
	inv := inventory.New(3, nil)
	result := item.NewStack(block.Anvil{Type: block.UndamagedAnvil()}, 1)
	if err := inv.SetItem(2, result); err != nil {
		t.Fatal(err)
	}
	menuMu.Lock()
	lastMenus[s] = Menu{container: ContainerAnvil{}, inventory: inv}
	menuMu.Unlock()
	t.Cleanup(func() {
		menuMu.Lock()
		delete(lastMenus, s)
		menuMu.Unlock()
	})

	tests := []struct {
		name string
		id   byte
		slot byte
		want byte
		idIn int32
	}{
		{name: "input", id: protocol.ContainerAnvilInput, slot: 1, want: 0},
		{name: "material", id: protocol.ContainerAnvilMaterial, slot: 2, want: 1},
		{name: "result preview", id: protocol.ContainerAnvilResultPreview, slot: 50, want: 2},
		{name: "created output", id: protocol.ContainerCreatedOutput, slot: 50, want: 2, idIn: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &protocol.TakeStackRequestAction{}
			a.Source = protocol.StackRequestSlotInfo{
				Container:      protocol.FullContainerName{ContainerID: test.id},
				Slot:           test.slot,
				StackNetworkID: test.idIn,
			}
			a.Destination = protocol.StackRequestSlotInfo{
				Container: protocol.FullContainerName{ContainerID: protocol.ContainerInventory},
			}

			updateActionContainerID(a, s)

			if got := a.Source.Container.ContainerID; got != protocol.ContainerLevelEntity {
				t.Fatalf("container ID = %d, want %d", got, protocol.ContainerLevelEntity)
			}
			if got := a.Source.Slot; got != test.want {
				t.Fatalf("slot = %d, want %d", got, test.want)
			}
			if test.idIn < 0 && a.Source.StackNetworkID != item_id(result) {
				t.Fatalf("stack network ID = %d, want %d", a.Source.StackNetworkID, item_id(result))
			}
		})
	}
}

func TestContainerAnvilRewritesDestroy(t *testing.T) {
	s := &session.Session{}
	menuMu.Lock()
	lastMenus[s] = Menu{container: ContainerAnvil{}, inventory: inventory.New(3, nil)}
	menuMu.Unlock()
	t.Cleanup(func() {
		menuMu.Lock()
		delete(lastMenus, s)
		menuMu.Unlock()
	})

	a := &protocol.DestroyStackRequestAction{Source: protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerAnvilMaterial},
		Slot:      2,
	}}
	updateActionContainerID(a, s)

	if got := a.Source.Container.ContainerID; got != protocol.ContainerLevelEntity {
		t.Fatalf("container ID = %d, want %d", got, protocol.ContainerLevelEntity)
	}
	if got := a.Source.Slot; got != 1 {
		t.Fatalf("slot = %d, want 1", got)
	}
}

func TestContainerAnvilSkipsVanillaCraftActions(t *testing.T) {
	s := &session.Session{}
	menuMu.Lock()
	lastMenus[s] = Menu{container: ContainerAnvil{}}
	menuMu.Unlock()
	t.Cleanup(func() {
		menuMu.Lock()
		delete(lastMenus, s)
		menuMu.Unlock()
	})

	requests := []protocol.ItemStackRequest{{
		Actions: []protocol.StackRequestAction{
			&protocol.CraftRecipeOptionalStackRequestAction{},
			&protocol.CreateStackRequestAction{},
			&protocol.ConsumeStackRequestAction{},
		},
	}}
	handleItemStackRequest(s, requests)

	if got := len(requests[0].Actions); got != 1 {
		t.Fatalf("action count = %d, want 1", got)
	}
	if _, ok := requests[0].Actions[0].(*protocol.ConsumeStackRequestAction); !ok {
		t.Fatalf("remaining action has type %T, want *protocol.ConsumeStackRequestAction", requests[0].Actions[0])
	}
}

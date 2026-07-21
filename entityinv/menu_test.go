package entityinv

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestEncodedTitle(t *testing.T) {
	tests := []struct {
		size   int
		prefix string
	}{
		{size: 1, prefix: "§1§0"},
		{size: 9, prefix: "§1§0"},
		{size: 10, prefix: "§2§0"},
		{size: 54, prefix: "§6§0"},
		{size: 55, prefix: "§6§1"},
	}
	for _, test := range tests {
		t.Run(test.prefix, func(t *testing.T) {
			title := encodedTitle("Menu", test.size)
			if !strings.HasPrefix(title, test.prefix) || !strings.HasSuffix(title, "Menu") {
				t.Fatalf("encodedTitle(%d) = %q", test.size, title)
			}
		})
	}
}

func TestAddActorIdentifier(t *testing.T) {
	data, err := nbt.Marshal(actorIdentifiers{IDList: []actorID{{ID: "minecraft:pig"}}})
	if err != nil {
		t.Fatal(err)
	}
	withActor := addActorIdentifier(data)
	if bytes.Equal(withActor, data) {
		t.Fatal("actor identifier was not added")
	}
	var actors actorIdentifiers
	if err := nbt.Unmarshal(withActor, &actors); err != nil {
		t.Fatal(err)
	}
	if got := actors.IDList[len(actors.IDList)-1].ID; got != actorIdentifier {
		t.Fatalf("last actor identifier = %q", got)
	}
	if duplicate := addActorIdentifier(withActor); !bytes.Equal(duplicate, withActor) {
		t.Fatal("actor identifier was added twice")
	}
}

func TestPack(t *testing.T) {
	if _, err := Pack(); err != nil {
		t.Fatal(err)
	}
}

package inv

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
)

type handler struct {
	inventory.NopHandler
	p    *player.Player
	menu Menu
}

func (h handler) HandleTake(ctx *event.Context, slot int, it item.Stack) {
	ctx.Cancel()
	h.menu.submittable.Submit(h.p, it)
}

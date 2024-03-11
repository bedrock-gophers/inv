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

// HandleTake ...
func (h handler) HandleTake(ctx *event.Context, _ int, it item.Stack) {
	ctx.Cancel()
	h.menu.submittable.Submit(h.p, it)
}

// HandlePlace ...
func (h handler) HandlePlace(ctx *event.Context, _ int, _ item.Stack) {
	ctx.Cancel()
}

// HandleDrop ...
func (h handler) HandleDrop(ctx *event.Context, _ int, it item.Stack) {
	ctx.Cancel()
	h.menu.submittable.Submit(h.p, it)
}

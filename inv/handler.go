package inv

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
)

// handler is only used for regular Menu customMenus allow their own handler.
type handler struct {
	inventory.NopHandler
	menu Menu
}

// HandleTake ...
func (h handler) HandleTake(ctx *inventory.Context, _ int, it item.Stack) {
	p := ctx.Val().(*player.Player)
	ctx.Cancel()
	h.menu.submittable.Submit(p, it)
}

// HandlePlace ...
func (h handler) HandlePlace(ctx *inventory.Context, _ int, _ item.Stack) {
	ctx.Cancel()
}

// HandleDrop ...
func (h handler) HandleDrop(ctx *inventory.Context, _ int, it item.Stack) {
	p := ctx.Val().(*player.Player)
	ctx.Cancel()
	h.menu.submittable.Submit(p, it)
}

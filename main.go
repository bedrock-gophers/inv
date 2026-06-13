package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/bedrock-gophers/inv/forminv"
	"github.com/bedrock-gophers/inv/inv"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
)

func main() {
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := server.DefaultConfig().Config(slog.Default())
	if err != nil {
		log.Fatalln(err)
	}
	conf.Listeners = intercept.WrapListeners(conf.Listeners)
	if err := forminv.AddToConfig(&conf); err != nil {
		log.Fatalln(err)
	}
	conf.ResourcesRequired = true

	srv := conf.New()
	srv.CloseOnProgramEnd()

	srv.Listen()
	for p := range srv.Accept() {
		p.Handle(playerHandler{})
		accept(p)
	}
}

type playerHandler struct {
	player.NopHandler
}

func (h playerHandler) HandleQuit(p *player.Player) {
	inv.CloseContainer(p) // should be called whenever a player leaves the server to prevent memory leaks
}

func accept(p *player.Player) {
	time.AfterFunc(1*time.Second, func() {
		//sub := MySubmittable{}

		//var stacks = make([]item.Stack, 54)
		//for i := 0; i < 54; i++ {
		//	stacks[i] = item.NewStack(block.StainedGlass{Colour: item.ColourRed()}, 1)
		//}

		//m := inv.NewMenu(sub, "test", inv.ContainerChest{DoubleChest: true}).WithStacks(stacks...)
		//inv.SendMenu(p, m)

		forminv.SendMenu(p, newExampleFormMenu(0))
	})
}

type MySubmittable struct{}

func (m MySubmittable) Submit(p *player.Player, it item.Stack) {
	fmt.Println("Submitted", it)
}

func (m MySubmittable) Close(p *player.Player) {
	fmt.Println("Closed")
}

type MyFormSubmittable struct{}

func (m MyFormSubmittable) Submit(p *player.Player, slot forminv.Slot) {
	fmt.Println("Submitted form slot", slot.Value)
	if count, ok := slot.Value.(int); ok {
		forminv.UpdateMenu(p, newExampleFormMenu(count+1))
	}
}

func (m MyFormSubmittable) Close(p *player.Player) {
	fmt.Println("Closed form inventory")
}

func newExampleFormMenu(count int) forminv.Menu {
	return forminv.NewMenu(MyFormSubmittable{}, fmt.Sprintf("form test #%d", count), forminv.LargeChest).WithSlots(
		forminv.NewSlot(10, fmt.Sprintf("Resend #%d", count+1), "textures/items/diamond_sword", count),
		forminv.NewSlot(13, "Ender Pearl", "textures/items/ender_pearl", "ender_pearl"),
		forminv.NewSlot(16, "Close", "textures/ui/cancel", "close"),
	)
}

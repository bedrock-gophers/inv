package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/bedrock-gophers/inv/entityinv"
	"github.com/bedrock-gophers/inv/inv"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
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
	if err := entityinv.AddToConfig(&conf); err != nil {
		log.Fatalln(err)
	}
	conf.ResourcesRequired = true

	srv := conf.New()
	intercept.Start(srv)
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
	entityinv.CloseContainer(p)
}

func accept(p *player.Player) {
	time.AfterFunc(1*time.Second, func() {
		stacks := make([]item.Stack, 59)
		for i := range stacks {
			stacks[i] = item.NewStack(block.StainedGlass{Colour: item.ColourRed()}, 1)
		}
		m := entityinv.NewMenu(MySubmittable{}, "59-slot Entity Inventory", len(stacks)).WithStacks(stacks...)
		entityinv.SendMenu(p, m)
	})
}

type MySubmittable struct{}

func (m MySubmittable) Submit(p *player.Player, it item.Stack) {
	fmt.Println("Submitted", it)
}

func (m MySubmittable) Close(p *player.Player) {
	fmt.Println("Closed")
}

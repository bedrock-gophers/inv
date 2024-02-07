package main

import (
	"fmt"
	"github.com/bedrock-gophers/invmenu/invmenu"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := server.DefaultConfig().Config(log)
	if err != nil {
		log.Fatalln(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	srv.Listen()
	invmenu.PlaceFakeChest(srv.World())

	for srv.Accept(accept) {

	}
}

func accept(p *player.Player) {
	time.AfterFunc(1*time.Second, func() {
		inv := inventory.New(27, func(slot int, before, after item.Stack) {})
		inv.Handle(h{})

		for i := range inv.Slots() {
			_ = inv.SetItem(i, item.NewStack(block.StainedGlassPane{
				Colour: item.ColourRed(),
			}, 1))
		}
		invmenu.ShowInventory(p, "", inv)
	})
}

type h struct {
	inventory.NopHandler
}

func (h) HandleTake(ctx *event.Context, slot int, it item.Stack) {
	fmt.Println(slot)
}

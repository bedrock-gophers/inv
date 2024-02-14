package main

import (
	"fmt"
	"github.com/bedrock-gophers/inv/inv"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sandertv/gophertunnel/minecraft/text"
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
	inv.PlaceFakeChest(srv.World(), cube.Pos{0, 255, 0})

	for srv.Accept(accept) {

	}
}

func accept(p *player.Player) {
	time.AfterFunc(1*time.Second, func() {
		in := inventory.New(27, func(slot int, before, after item.Stack) {})
		in.Handle(h{})

		for i := range in.Slots() {
			_ = in.SetItem(i, item.NewStack(block.StainedGlassPane{
				Colour: item.ColourRed(),
			}, 1))
		}
		inv.ShowMenu(p, in, text.Colourf("<red>Test</red>"))
	})
}

type h struct {
	inventory.NopHandler
}

func (h) HandleTake(ctx *event.Context, slot int, it item.Stack) {
	fmt.Println(slot)
}

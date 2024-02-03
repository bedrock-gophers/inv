package main

import (
	"github.com/bedrock-gophers/invmenu/invmenu"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
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

	for srv.Accept(accept) {

	}
}

func accept(p *player.Player) {
	time.AfterFunc(1*time.Second, func() {
		inv := inventory.New(27, func(slot int, before, after item.Stack) {})

		for i := range inv.Slots() {
			_ = inv.SetItem(i, item.NewStack(block.StainedGlassPane{
				Colour: item.ColourRed(),
			}, 1))
		}
		invmenu.ShowInventory(p, "", inv)
	})
}

package main

import (
	"github.com/bedrock-gophers/inv/inv"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block/cube"
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
	inv.PlaceFakeContainer(srv.World(), cube.Pos{0, 255, 0})

	for srv.Accept(accept) {

	}
}

func accept(p *player.Player) {
	inv.RedirectPlayerPackets(p)
	time.AfterFunc(1*time.Second, func() {
		testInv := inventory.New(54, func(slot int, before, after item.Stack) {

		})
		_, _ = testInv.AddItem(item.NewStack(item.Diamond{}, 1))
		testInv.Handle(inventory.NopHandler{})

		m := inv.NewCustomMenu("test", inv.ChestContainer{DoubleChest: true}, testInv)
		inv.SendMenu(p, m)
	})
}

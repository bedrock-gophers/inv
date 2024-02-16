package main

import (
	"fmt"
	"github.com/bedrock-gophers/inv/inv"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	fmt.Println(block.Chest{}.Hash())
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
		sub := MySubmittable{}

		var stacks = make([]item.Stack, 27)
		for i := 0; i < 27; i++ {
			stacks[i] = item.NewStack(block.StainedGlass{Colour: item.ColourRed()}, 1)
		}

		m := inv.NewMenu(sub, "test").WithStacks(stacks...)
		inv.SendMenu(p, m)

		time.AfterFunc(1*time.Second, func() {
			inv.SendMenu(p, inv.NewMenu(MySubmittable{}, "test").WithStacks(stacks...))
		})
	})
}

type MySubmittable struct{}

func (m MySubmittable) Submit(p *player.Player, it item.Stack) {
	fmt.Println("Submitted", it)
}

func (m MySubmittable) Close(p *player.Player) {
	fmt.Println("Closed")
}

package inv

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
)

// barrel is a block.ContainerViewer that can be used to view a container.
type barrel struct {
	block.Barrel
}

// RemoveViewer removes a viewer from a container. The viewer is removed from the container at the position
func (barrel) RemoveViewer(v block.ContainerViewer, _ *world.World, _ cube.Pos) {
	CloseContainer(containerViewerToPlayer(v))
}

// chest is a block.ContainerViewer that can be used to view a container.
type chest struct {
	block.Chest
}

// RemoveViewer removes a viewer from a container. The viewer is removed from the container at the position
func (chest) RemoveViewer(v block.ContainerViewer, _ *world.World, _ cube.Pos) {
	CloseContainer(containerViewerToPlayer(v))
}

func containerViewerToPlayer(v block.ContainerViewer) *player.Player {
	return v.(*session.Session).Controllable().(*player.Player)
}
